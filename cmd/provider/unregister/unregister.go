package unregister

import (
	"context"
	"errors"
	"io"
	"strconv"

	providerv1grpc "buf.build/gen/go/mpapenbr/testrepo/grpc/go/testrepo/provider/v1/providerv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/testrepo/protocolbuffers/go/testrepo/event/v1"
	providerv1 "buf.build/gen/go/mpapenbr/testrepo/protocolbuffers/go/testrepo/provider/v1"
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
	conn, err := util.ConnectGRPC(config.DefaultCliArgs().Addr)
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
		return err
	}
	defer conn.Close()
	req := providerv1.UnregisterEventRequest{
		EventSelector: resolveEvent(arg),
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
	conn, err := util.ConnectGRPC(config.DefaultCliArgs().Addr)
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
	for {
		resp, err := r.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Error("error fetching events", log.ErrorField(err))
			break
		} else {
			log.Debug("got event: ",
				log.Uint32("id", resp.Event.Id),
				log.String("key", resp.Event.Key))
		}
	}

	return nil
}

//nolint:gosec //check is not needed here
func resolveEvent(arg string) *eventv1.EventSelector {
	if id, err := strconv.Atoi(arg); err == nil {
		return &eventv1.EventSelector{Arg: &eventv1.EventSelector_Id{Id: int32(id)}}
	}
	return &eventv1.EventSelector{Arg: &eventv1.EventSelector_Key{Key: arg}}
}
