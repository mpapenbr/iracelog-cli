package list

import (
	"context"
	"errors"
	"io"

	providerv1grpc "buf.build/gen/go/mpapenbr/testrepo/grpc/go/testrepo/provider/v1/providerv1grpc"
	providerv1 "buf.build/gen/go/mpapenbr/testrepo/protocolbuffers/go/testrepo/provider/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewProviderListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists current data provider.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return listEvents()
		},
	}
	return cmd
}

func listEvents() error {
	log.Debug("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGRPC(config.DefaultCliArgs().Addr)
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
		return err
	}
	defer conn.Close()
	req := providerv1.ListLiveEventsRequest{}
	c := providerv1grpc.NewProviderServiceClient(conn)
	r, err := c.ListLiveEvents(context.Background(), &req)
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
			log.Debug("got event: ", log.Any("event", resp))
		}
	}
	return nil
}
