package tire

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/mpapenbr/iracelog-cli/cmd/event/check/options"
	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

var carNumFilter []string

func NewCheckTireCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tire",
		Short: "checks if tire compound changes when not in pit",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkTires(cmd.Context(), args[0])
		},
	}
	cmd.Flags().StringSliceVar(&carNumFilter, "filter-carnum", []string{},
		"filter cars by car number")
	return cmd
}

//nolint:funlen,gocyclo // by design
func checkTires(ctx context.Context, arg string) {
	logger := log.GetFromContext(ctx)
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	var startSel *commonv1.StartSelector
	if startSel, err = util.ResolveStartSelector(
		options.SessionTime,
		options.RecordStamp); err != nil {
		logger.Error("could not resolve start selector",
			log.ErrorField(err),
			log.Duration("session-time", options.SessionTime),
			log.String("record-stamp", options.RecordStamp))
		return
	}
	logger.Info("start selector resolved", log.Any("start-selector", startSel))

	var eventData *eventv1.GetEventResponse
	if eventData, err = loadEvent(conn, arg); err != nil {
		logger.Error("could not load event", log.ErrorField(err), log.String("event", arg))
		return
	}
	carNumByIdx := make(map[int32]string)
	carIdxByNum := make(map[string]int32)
	for _, c := range eventData.Car.Entries {
		carNumByIdx[int32(c.Car.CarIdx)] = c.Car.CarNumber
		carIdxByNum[c.Car.CarNumber] = int32(c.Car.CarIdx)
	}

	req := racestatev1.GetStateStreamRequest{
		Event: util.ResolveEvent(arg),
		Start: startSel,
		Num:   options.NumEntries,
	}

	c := racestatev1grpc.NewRaceStateServiceClient(conn)

	type tcChange struct {
		SessionTime float32
		TcOld       uint32
		TcNew       uint32
		Lap         int32
		StintLap    uint32
		TrackPos    float32
	}
	type tcInfo struct {
		tc      uint32 // last seen tire compound
		changes []*tcChange
	}
	tc := make(map[int32]*tcInfo)

	var resp grpc.ServerStreamingClient[racestatev1.GetStateStreamResponse]

	if resp, err = c.GetStateStream(ctx, &req); err != nil {
		logger.Error("could not load states for event",
			log.ErrorField(err),
			log.String("event", arg))
		return
	}

	for {
		sr, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			log.Debug("EOF")
			break
		}
		if err != nil {
			logger.Error("error fetching states", log.ErrorField(err))
			return
		}
		s := sr.GetState()
		for _, c := range s.Cars {
			if tcData, ok := tc[c.CarIdx]; ok {
				//nolint:exhaustive // by design
				switch c.State {
				case racestatev1.CarState_CAR_STATE_PIT:
					tcData.tc = c.TireCompound.RawValue
				case racestatev1.CarState_CAR_STATE_RUN:
					if c.TireCompound.RawValue != tcData.tc {
						logger.Debug("Tire compound changed",
							log.Int("carIdx", int(c.CarIdx)),
							log.Uint32("old", tcData.tc),
							log.Uint32("now", c.TireCompound.RawValue),
							log.Float32("time", s.Session.SessionTime),
						)
						tcData.changes = append(tcData.changes, &tcChange{
							SessionTime: s.Session.SessionTime,
							TcOld:       tcData.tc,
							TcNew:       c.TireCompound.RawValue,
							Lap:         c.Lap,
							StintLap:    c.StintLap,
							TrackPos:    c.TrackPos,
						})
						tcData.tc = c.TireCompound.RawValue
					}
				}
			} else {
				tcData = &tcInfo{tc: c.TireCompound.RawValue, changes: []*tcChange{}}
				tc[c.CarIdx] = tcData
			}
		}
	}

	sortedCarNum := slices.Sorted(maps.Values(carNumByIdx))
	for _, cn := range sortedCarNum {
		carIdx := carIdxByNum[cn]
		tcData := tc[carIdx]
		if len(carNumFilter) > 0 {
			if !slices.Contains(carNumFilter, cn) {
				continue
			}
		}
		fmt.Printf("Car %s (carIdx:%d)\n", cn, carIdx)
		for _, c := range tcData.changes {
			fmt.Printf("Session %.1f: %d -> %d (lap %d) StintLap: %d TrackPos: %.4f\n",
				c.SessionTime, c.TcOld, c.TcNew, c.Lap, c.StintLap, c.TrackPos)
		}
	}
}

//nolint:whitespace // editor/linter issue
func loadEvent(conn *grpc.ClientConn, arg string) (
	ret *eventv1.GetEventResponse, err error,
) {
	c := eventv1grpc.NewEventServiceClient(conn)
	req := eventv1.GetEventRequest{
		EventSelector: util.ResolveEvent(arg),
	}

	if ret, err = c.GetEvent(context.Background(), &req); err != nil {
		return nil, err
	}
	return ret, nil
}
