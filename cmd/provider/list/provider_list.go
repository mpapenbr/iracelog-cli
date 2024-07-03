package list

import (
	"context"

	providerv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/provider/v1/providerv1grpc"
	providerv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/provider/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewProviderListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists current data provider.",

		RunE: func(cmd *cobra.Command, args []string) error {
			return listEvents()
		},
	}
	return cmd
}

func listEvents() error {
	logger := log.GetLoggerManager().GetDefaultLogger()
	logger.Debug("connect grpc server ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
		return err
	}
	defer conn.Close()
	req := providerv1.ListLiveEventsRequest{}
	c := providerv1grpc.NewProviderServiceClient(conn)
	r, err := c.ListLiveEvents(context.Background(), &req)
	if err != nil {
		logger.Error("could not get events", log.ErrorField(err))
		return err
	}

	for i := range r.Events {
		logger.Debug("got event: ", log.Any("event", r.Events[i]))
	}
	return nil
}
