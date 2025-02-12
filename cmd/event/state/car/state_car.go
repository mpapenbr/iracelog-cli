package car

import (
	"context"
	"math"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/mpapenbr/iracelog-cli/cmd/event/state/options"
	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
	"github.com/mpapenbr/iracelog-cli/util/output"
	"github.com/mpapenbr/iracelog-cli/util/output/car"
)

var (
	attrs  []string
	format string
	carNum string
)

func NewStateCarCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "car",
		Short: "shows state data for a car",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showCarData(cmd.Context(), args[0])
		},
	}

	cmd.Flags().StringVar(&format, "format", "text",
		"output format (text, json,csv)")
	cmd.Flags().StringSliceVar(&attrs, "attrs", []string{},
		"session attributes to display")
	cmd.Flags().StringVar(&carNum, "carnum", "",
		"filter data for this car")
	//nolint:errcheck // by design
	cmd.MarkFlagRequired("carnum")
	return cmd
}

//nolint:funlen,gocyclo // by design
func showCarData(ctx context.Context, arg string) {
	logger := log.GetFromContext(ctx)
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()

	var startSel *commonv1.StartSelector
	if startSel, err = util.ResolveStartSelector2(
		options.BuildStartSelParam()); err != nil {
		logger.Error("could not resolve start selector",
			log.ErrorField(err),
			log.Duration("session-time", options.SessionTime),
			log.String("record-stamp", options.RecordStamp))
		return
	}

	var eventData *eventv1.GetEventResponse
	if eventData, err = loadEvent(conn, arg); err != nil {
		logger.Error("could not load event", log.ErrorField(err), log.String("event", arg))
		return
	}
	var carIdx int32 = -1
	for _, ce := range eventData.Car.Entries {
		if ce.Car.CarNumber == carNum {
			carIdx = int32(ce.Car.CarIdx)
			break
		}
	}
	if carIdx == -1 {
		logger.Error("could not find car", log.String("carNum", carNum))
		return
	}
	remain := math.MaxInt32
	if options.NumEntries > 0 {
		remain = int(options.NumEntries)
	}
	toFetchEntries := func() int32 {
		if options.NumEntries > 0 {
			return min(min(500, options.NumEntries), int32(remain))
		}
		return int32(500)
	}

	req := racestatev1.GetStatesRequest{
		Event: util.ResolveEvent(arg),
		Start: startSel,
		Num:   toFetchEntries(),
	}

	c := racestatev1grpc.NewRaceStateServiceClient(conn)
	var resp *racestatev1.GetStatesResponse
	if resp, err = c.GetStates(ctx, &req); err != nil {
		logger.Error("could not load states for event",
			log.ErrorField(err),
			log.String("event", arg))
		return
	}
	logger.Debug("States loaded.", log.Int("num", len(resp.States)))

	opts := []car.Option{}

	if format != "" {
		if f, errFmt := output.ParseFormat(format); errFmt == nil {
			opts = append(opts, car.WithFormat(f))
		}
	}
	if len(attrs) > 0 {
		carAttrs := []car.CarAttr{}
		for _, c := range attrs {
			v, _ := car.ParseCarAttr(c)
			carAttrs = append(carAttrs, v)
		}
		opts = append(opts, car.WithCarAttrs(carAttrs))
	} else {
		opts = append(opts, car.WithAllCarAttrs())
	}
	out := car.NewCarOutput(opts...)
	out.Header()
	for {
		var resp *racestatev1.GetStatesResponse
		if resp, err = c.GetStates(ctx, &req); err != nil {
			logger.Error("could not load states for event",
				log.ErrorField(err),
				log.String("event", arg))
			return
		}
		if len(resp.States) == 0 {
			break
		}
		for _, s := range resp.States {
			for idx, c := range s.Cars {
				if c.CarIdx == carIdx {
					out.Line(s.Session, s.Cars[idx])
					break
				}
			}
		}
		if options.NumEntries > 0 {
			remain -= len(resp.States)
		}
		logger.Debug("States loaded.",
			log.Int("num", len(resp.States)),
			log.Int("remain", remain))
		if remain <= 0 {
			break
		}
		req.Start = &commonv1.StartSelector{
			Arg: &commonv1.StartSelector_Id{
				Id: resp.GetLastId() + 1,
			},
		}
		req.SetNum(toFetchEntries())
	}
	out.Flush()
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
