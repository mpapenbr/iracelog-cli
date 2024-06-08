package register

import (
	"context"

	providerv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/provider/v1/providerv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	providerv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/provider/v1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewProviderRegisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "registers a data provider (for debugging only!).",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return registerEvent()
		},
	}
	cmd.PersistentFlags().StringVarP(&config.DefaultCliArgs().Token,
		"token", "t", "", "authentication token")
	return cmd
}

func registerEvent() error {
	log.Debug("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
		return err
	}
	defer conn.Close()
	req := providerv1.RegisterEventRequest{
		Event: &eventv1.Event{Key: uuid.New().String()},
	}
	c := providerv1grpc.NewProviderServiceClient(conn)
	md := metadata.Pairs("api-token", config.DefaultCliArgs().Token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	r, err := c.RegisterEvent(ctx, &req)
	if err != nil {
		log.Error("could not get events", log.ErrorField(err))
		return err
	}
	log.Debug("got event: ", log.Any("event", r))

	return nil
}
