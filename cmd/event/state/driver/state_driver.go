package driver

import (
	"context"
	"errors"
	"io"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/mpapenbr/iracelog-cli/cmd/event/state/options"
	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

var (
	attrs  []string
	format string
)

func NewStateDriverDataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "driver",
		Short: "shows state driver data",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showDriverData(cmd.Context(), args[0])
		},
	}

	cmd.Flags().StringVar(&format, "format", "text",
		"output format (text, json,csv)")
	cmd.Flags().StringSliceVar(&attrs, "attrs", []string{},
		"session attributes to display")

	return cmd
}

//nolint:funlen // by design
func showDriverData(ctx context.Context, arg string) {
	logger := log.GetFromContext(ctx)
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	var startSel *commonv1.StartSelector
	if startSel, err = util.ResolveStartSelector2(
		options.BuildStartSelParam(),
	); err != nil {
		logger.Error("could not resolve start selector",
			log.ErrorField(err),
			log.Duration("session-time", options.SessionTime),
			log.String("record-stamp", options.RecordStamp))
		return
	}
	logger.Info("start selector resolved", log.Any("start-selector", startSel))

	req := racestatev1.GetDriverDataStreamRequest{
		Event: util.ResolveEvent(arg),
		Start: startSel,
		Num:   options.NumEntries,
	}

	c := racestatev1grpc.NewRaceStateServiceClient(conn)

	var resp grpc.ServerStreamingClient[racestatev1.GetDriverDataStreamResponse]

	if resp, err = c.GetDriverDataStream(ctx, &req); err != nil {
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
			logger.Error("error fetching driver data", log.ErrorField(err))
			return
		}

		logger.Debug("data",
			log.Uint32("sessionNum", s.DriverData.SessionNum),
			log.Float32("sessionTime", s.DriverData.SessionTime),
			log.Time("recordStamp", s.DriverData.Timestamp.AsTime()),
		)
	}
}
