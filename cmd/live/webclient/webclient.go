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

func NewLiveWebclientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webclient",
		Short: "simulates a webclient for live data",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			liveWebclient(args[0])
		},
	}

	return cmd
}

func liveWebclient(eventArg string) {
	logger := log.GetLoggerManager().GetDefaultLogger()

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts = append(opts, simulate.WithContext(ctx))
	wc := simulate.NewWebclient(opts...)
	//nolint:errcheck // by design
	wc.Start(sel)
}
