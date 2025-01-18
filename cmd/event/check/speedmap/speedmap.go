package speedmap

import (
	"context"
	"math"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/mpapenbr/iracelog-cli/cmd/event/check/options"
	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewCheckSpeedmapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "speedmap",
		Short: "Commands to check speedmap data consistency.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkSpeedmaps(cmd.Context(), args[0])
		},
	}
	return cmd
}

//nolint:funlen // by design
func checkSpeedmaps(ctx context.Context, arg string) {
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
	req := racestatev1.GetSpeedmapsRequest{
		Event: util.ResolveEvent(arg),
		Start: startSel,
		Num:   toFetchEntries(),
	}

	c := racestatev1grpc.NewRaceStateServiceClient(conn)
	var prevTs *timestamppb.Timestamp = nil
	for {
		var resp *racestatev1.GetSpeedmapsResponse

		if resp, err = c.GetSpeedmaps(ctx, &req); err != nil {
			logger.Error("could not load speedmaps for event",
				log.ErrorField(err),
				log.String("event", arg))
			return
		}
		if len(resp.Speedmaps) == 0 {
			break
		}
		for _, s := range resp.Speedmaps {
			if prevTs != nil {
				delta := s.Timestamp.AsTime().Sub(prevTs.AsTime())
				if delta > options.GapThreshold {
					logger.Info("Gap detected.",
						log.Time("prev", prevTs.AsTime()),
						log.Time("this", s.Timestamp.AsTime()),

						log.Duration("delta", delta))
				}
			}
			prevTs = s.Timestamp
		}
		if options.NumEntries > 0 {
			remain -= len(resp.Speedmaps)
		}
		logger.Debug("Speedmaps loaded.",
			log.Int("num", len(resp.Speedmaps)),
			log.Int("remain", remain))
		req.Start = &commonv1.StartSelector{
			Arg: &commonv1.StartSelector_RecordStamp{
				RecordStamp: resp.LastTs,
			},
		}
		req.SetNum(toFetchEntries())
	}
}
