package analysis

import (
	"context"
	"errors"
	"io"

	livedatav1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/livedata/v1/livedatav1grpc"
	livedatav1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/livedata/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewLiveCarOccupancyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "caroccupancy",
		Short: "receives live car occupancy data",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			liveCarInfo(args[0])
		},
	}
	return cmd
}

func liveCarInfo(eventArg string) {
	sel := util.ResolveEvent(eventArg)
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Error("did not connect", log.ErrorField(err))
		return
	}
	defer conn.Close()
	req := livedatav1.LiveCarOccupanciesRequest{
		Event: sel,
	}
	c := livedatav1grpc.NewLiveDataServiceClient(conn)
	r, err := c.LiveCarOccupancies(context.Background(), &req)
	if err != nil {
		log.Error("could not get live data", log.ErrorField(err))
		return
	}

	for {
		resp, err := r.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Error("error fetching live state", log.ErrorField(err))
			return
		} else {
			log.Debug("got count: ", log.Int("count", len(resp.CarOccupancies)))
		}
	}
}
