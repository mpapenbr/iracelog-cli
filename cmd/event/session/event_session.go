package session

import (
	"context"
	"time"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
	"github.com/mpapenbr/iracelog-cli/util/output"
	"github.com/mpapenbr/iracelog-cli/util/output/session"
)

var (
	sessionTime time.Duration
	recordStamp string
	num         int

	attrs  []string
	format string
)

func NewEventSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "shows session data for an event",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			showSessionData(cmd.Context(), args[0])
		},
	}

	cmd.Flags().DurationVar(&sessionTime, "session-time", 0,
		"session time as duration where data should begin (for example: 10m)")
	cmd.Flags().StringVar(&recordStamp, "record-stamp", "",
		"timestamp time where data should begin")
	cmd.Flags().IntVar(&num, "num", 20,
		"number of entries to show")
	cmd.MarkFlagsMutuallyExclusive("session-time", "record-stamp")
	cmd.MarkFlagsOneRequired("session-time", "record-stamp")
	cmd.Flags().StringSliceVar(&attrs, "attrs", []string{},
		"session attributes to display")
	cmd.Flags().StringVar(&format, "format", "text",
		"output format (text, json,csv)")

	return cmd
}

//nolint:funlen // by design
func showSessionData(ctx context.Context, arg string) {
	logger := log.GetFromContext(ctx)
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	var startSel *commonv1.StartSelector
	if startSel, err = util.ResolveStartSelector(sessionTime, recordStamp); err != nil {
		logger.Error("could not resolve start selector",
			log.ErrorField(err),
			log.Duration("session-time", sessionTime),
			log.String("record-stamp", recordStamp))
		return
	}

	req := racestatev1.GetStatesRequest{
		Event: util.ResolveEvent(arg),
		Start: startSel,
		Num:   int32(num),
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

	opts := []session.Option{}

	if format != "" {
		if f, err := output.ParseFormat(format); err == nil {
			opts = append(opts, session.WithFormat(f))
		}
	}
	if len(attrs) > 0 {
		sessionAttrs := []session.SessionAttr{}
		for _, c := range attrs {
			v, _ := session.ParseSessionAttr(c)
			sessionAttrs = append(sessionAttrs, v)
		}
		opts = append(opts, session.WithSessionAttrs(sessionAttrs))
	} else {
		opts = append(opts, session.WithAllSessionAttrs())
	}
	out := session.NewSessionOutput(opts...)
	out.Header()
	for i := range resp.States {
		s := resp.States[i]
		out.Line(s)
	}
	out.Flush()
}
