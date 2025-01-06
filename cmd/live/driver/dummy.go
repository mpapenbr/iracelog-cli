package driver

import (
	"context"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewSendEmptyDriverData() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "empty",
		Short: "send empty driver data request",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			sendEmptyDriverData(args[0])
		},
	}
	cmd.PersistentFlags().StringVarP(&config.DefaultCliArgs().Token,
		"token", "t", "", "authentication token")
	return cmd
}

func sendEmptyDriverData(eventArg string) {
	sel := util.ResolveEvent(eventArg)
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Error("did not connect", log.ErrorField(err))
		return
	}
	defer conn.Close()
	req := racestatev1.PublishDriverDataRequest{
		Event:     sel,
		Timestamp: timestamppb.Now(),
	}
	c := racestatev1grpc.NewRaceStateServiceClient(conn)
	md := metadata.Pairs("api-token", config.DefaultCliArgs().Token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	_, err = c.PublishDriverData(ctx, &req)
	if err != nil {
		log.Error("could send drviver data", log.ErrorField(err))
		return
	}
}
