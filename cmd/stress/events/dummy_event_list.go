package events

import (
	"context"
	"errors"
	"io"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/mpapenbr/iracelog-cli/cmd/stress/config"
	appCfg "github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	myStress "github.com/mpapenbr/iracelog-cli/stress"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewStressDummyEventListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "dummy to test stress component ",
		Run: func(cmd *cobra.Command, args []string) {
			experimental()
		},
	}

	return cmd
}

func experimental() {
	configOptions := config.CollectStandardJobProcessorOptions()
	configOptions = append(configOptions,
		myStress.WithClientProvider(func() *grpc.ClientConn {
			c, err := util.ConnectGrpc(appCfg.DefaultCliArgs())
			if err != nil {
				log.Fatal("could  not connect server", log.ErrorField(err))
			}
			log.Debug("connected to server")
			return c
		}),
		myStress.WithJobHandler(func(j *myStress.Job) error {
			req := eventv1.GetEventsRequest{}
			c := eventv1grpc.NewEventServiceClient(j.Client)
			r, err := c.GetEvents(context.Background(), &req)
			if err != nil {
				log.Error("could not get events", log.ErrorField(err))
				return err
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
			return nil
		}),
	)
	jobProcessor := myStress.NewJobProcessor(configOptions...)
	jobProcessor.Run()
}
