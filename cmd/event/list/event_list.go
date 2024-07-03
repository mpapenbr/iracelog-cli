package list

import (
	"context"
	"errors"
	"io"

	eventv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewEventListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists stored events.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			listEvents()
		},
	}
	return cmd
}

func listEvents() {
	logger := log.GetLoggerManager().GetDefaultLogger()
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	req := eventv1.GetEventsRequest{}
	c := eventv1grpc.NewEventServiceClient(conn)
	r, err := c.GetEvents(context.Background(), &req)
	if err != nil {
		logger.Error("could not get events", log.ErrorField(err))
		return
	}

	for {
		resp, err := r.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			logger.Error("error fetching events", log.ErrorField(err))
			break
		} else {
			logger.Info("got event: ",
				log.Uint32("id", resp.Event.Id),
				log.String("key", resp.Event.Key))
		}
	}
}
