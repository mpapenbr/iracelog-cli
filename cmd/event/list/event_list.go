package list

import (
	"context"
	"errors"
	"io"

	eventv1grpc "buf.build/gen/go/mpapenbr/testrepo/grpc/go/testrepo/event/v1/eventv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/testrepo/protocolbuffers/go/testrepo/event/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
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
	log.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := grpc.Dial(config.DefaultCliArgs().Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	req := eventv1.GetEventsRequest{}
	c := eventv1grpc.NewEventServiceClient(conn)
	r, err := c.GetEvents(context.Background(), &req)
	if err != nil {
		log.Error("could not get events", log.ErrorField(err))
		return
	}

	for {
		resp, err := r.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Error("error fetching events", log.ErrorField(err))
			break
		} else {
			log.Info("got event: ",
				log.Uint32("id", resp.Event.Id),
				log.String("key", resp.Event.Key))
		}
	}
}
