package unregister

import (
	"context"

	providerv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/provider/v1/providerv1grpc"
	providerv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/provider/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewProviderUnregisterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregister",
		Short: "unregisters a data provider ",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return unregisterEvent(args[0])
		},
	}
	cmd.PersistentFlags().StringVarP(&config.DefaultCliArgs().Token,
		"token", "t", "", "authentication token")
	return cmd
}

func NewProviderUnregisterAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unregisterAll",
		Short: "unregisters all data providers ",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			return unregisterAllEvents()
		},
	}
	cmd.PersistentFlags().StringVarP(&config.DefaultCliArgs().Token,
		"token", "t", "", "authentication token")
	return cmd
}

func unregisterEvent(arg string) error {
	log.Debug("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
		return err
	}
	defer conn.Close()
	req := providerv1.UnregisterEventRequest{
		EventSelector: util.ResolveEvent(arg),
	}
	c := providerv1grpc.NewProviderServiceClient(conn)
	md := metadata.Pairs("api-token", config.DefaultCliArgs().Token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	r, err := c.UnregisterEvent(ctx, &req)
	if err != nil {
		log.Error("could not unregister event", log.ErrorField(err))
		return err
	}
	log.Info("got event: ", log.Any("event", r))

	return nil
}

func unregisterAllEvents() error {
	log.Debug("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
		return err
	}
	defer conn.Close()
	req := providerv1.UnregisterAllRequest{}
	c := providerv1grpc.NewProviderServiceClient(conn)
	md := metadata.Pairs("api-token", config.DefaultCliArgs().Token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	r, err := c.UnregisterAll(ctx, &req)
	if err != nil {
		log.Error("could not unregister all", log.ErrorField(err))
		return err
	}

	for i := range r.Events {
		log.Debug("got event: ",
			log.Uint32("id", r.Events[i].Event.Id),
			log.String("key", r.Events[i].Event.Key))
	}

	return nil
}
