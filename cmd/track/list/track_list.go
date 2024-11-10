package list

import (
	"context"
	"errors"
	"io"
	"time"

	trackv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/track/v1/trackv1grpc"
	trackv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/track/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
	"github.com/mpapenbr/iracelog-cli/util/output"
	"github.com/mpapenbr/iracelog-cli/util/output/track"
)

var (
	attrs  []string
	format string
)

func NewTrackListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists available tracks.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			listTracks(cmd.Context())
		},
	}
	cmd.Flags().StringSliceVar(&attrs, "attrs", []string{},
		"session attributes to display")
	cmd.Flags().StringVar(&format, "format", "text",
		"output format (text, json,csv)")

	return cmd
}

//nolint:funlen // by design
func listTracks(ctx context.Context) {
	logger := log.GetFromContext(ctx)
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	req := trackv1.GetTracksRequest{}
	c := trackv1grpc.NewTrackServiceClient(conn)
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	r, err := c.GetTracks(reqCtx, &req)
	if err != nil {
		logger.Error("could not get tracks", log.ErrorField(err))
		return
	}

	opts := []track.Option{}

	if format != "" {
		if f, err := output.ParseFormat(format); err == nil {
			opts = append(opts, track.WithFormat(f))
		}
	}
	if len(attrs) > 0 {
		trackAttrs := []track.TrackAttr{}
		for _, c := range attrs {
			v, _ := track.ParseTrackAttr(c)
			trackAttrs = append(trackAttrs, v)
		}
		opts = append(opts, track.WithTrackAttrs(trackAttrs))
	} else {
		opts = append(opts, track.WithAllTrackAttrs())
	}
	out := track.NewTrackOutput(opts...)
	out.Header()

	for {
		resp, err := r.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			logger.Error("error fetching track", log.ErrorField(err))
			break
		} else {
			out.Line(resp.Track)
		}
	}
	out.Flush()
}
