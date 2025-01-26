package state

import (
	"context"
	"errors"
	"io"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/mpapenbr/iracelog-cli/cmd/event/check/options"
	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewCheckStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Commands to check state data consistency.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			checkStatesStream(cmd.Context(), args[0])
		},
	}
	return cmd
}

//nolint:funlen // by design
func checkStatesStream(ctx context.Context, arg string) {
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

	req := racestatev1.GetStateStreamRequest{
		Event: util.ResolveEvent(arg),
		Start: startSel,
		Num:   options.NumEntries,
	}

	c := racestatev1grpc.NewRaceStateServiceClient(conn)
	var prevTs *timestamppb.Timestamp = nil

	var resp grpc.ServerStreamingClient[racestatev1.GetStateStreamResponse]

	if resp, err = c.GetStateStream(ctx, &req); err != nil {
		logger.Error("could not load states for event",
			log.ErrorField(err),
			log.String("event", arg))
		return
	}

	for {
		s, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			log.Debug("EOF")
			break
		}
		if err != nil {
			logger.Error("error fetching states", log.ErrorField(err))
			return
		}
		if prevTs != nil {
			delta := s.State.Timestamp.AsTime().Sub(prevTs.AsTime())
			if delta > options.GapThreshold {
				logger.Info("Gap detected.",
					log.Time("prev", prevTs.AsTime()),
					log.Time("this", s.State.Timestamp.AsTime()),
					log.Float32("thisSessionTime", s.State.Session.SessionTime),
					log.Duration("delta", delta))
			}
		}
		prevTs = s.State.Timestamp
	}
}
