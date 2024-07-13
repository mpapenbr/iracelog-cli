package webclient

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
	"github.com/mpapenbr/iracelog-cli/util/simulate"
)

var (
	durationArg string
	statsArg    string
)

func NewLiveWebclientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webclient",
		Short: "simulates a webclient for live data",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			liveWebclient(cmd.Context(), args[0])
		},
	}
	cmd.Flags().StringVarP(&durationArg,
		"duration", "d", "", "simulate client for this duration")

	cmd.Flags().StringVar(&statsArg,
		"stats", "", "print stats with this interval duration  example: \"10s\"")
	return cmd
}

func liveWebclient(mainCtx context.Context, eventArg string) {
	logger := log.GetFromContext(mainCtx)

	sel := util.ResolveEvent(eventArg)
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Error("did not connect", log.ErrorField(err))
		return
	}
	defer conn.Close()

	opts := []simulate.Option{
		simulate.WithClient(conn),
	}
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	if durationArg != "" {
		if d, err := time.ParseDuration(durationArg); err == nil {
			ctx, cancel = context.WithTimeout(context.Background(), d)
		}
	}
	defer cancel()

	opts = append(opts, simulate.WithContext(ctx))
	if statsArg != "" {
		if d, err := time.ParseDuration(statsArg); err == nil {
			opts = append(opts, simulate.WithStatsCallback(d, func(s *simulate.Stats) {
				logger.Info("stats", log.Any("stats", s))
			}))
		}
	}

	wc := simulate.NewWebclient(opts...)
	//nolint:errcheck // by design
	wc.Start(sel)
}
