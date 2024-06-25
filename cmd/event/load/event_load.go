package load

import (
	"context"

	eventv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewEventLoadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "load stored event.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			loadEvent(args[0])
		},
	}

	return cmd
}

func loadEvent(arg string) {
	log.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()

	req := eventv1.GetEventRequest{
		EventSelector: util.ResolveEvent(arg),
	}

	ctx := context.Background()
	c := eventv1grpc.NewEventServiceClient(conn)
	if _, err := c.GetEvent(ctx, &req); err != nil {
		log.Error("could not load event", log.ErrorField(err), log.String("event", arg))
		return
	}
	log.Info("Event loaded.")
}
