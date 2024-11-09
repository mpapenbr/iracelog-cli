package edit

import (
	"context"
	"time"

	eventv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

var (
	editName           string
	editDescription    string
	editKey            string
	editMinSessionTime string
	editMaxSessionTime string
)

func NewEventEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "edit selected event data.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			editEvent(args[0])
		},
	}
	cmd.Flags().StringVarP(&config.DefaultCliArgs().Token,
		"token", "t", "", "authentication token")
	cmd.Flags().StringVarP(&editName,
		"name", "n", "", "new event name")
	cmd.Flags().StringVarP(&editDescription,
		"description", "d", "", "new event description")
	cmd.Flags().StringVarP(&editKey,
		"key", "k", "", "new event key (be very careful with this)")
	cmd.Flags().StringVar(&editMinSessionTime,
		"min-session-time", "", "minimum session time as duration")
	cmd.Flags().StringVar(&editMaxSessionTime,
		"max-session-time", "", "maximum session time as duration")
	return cmd
}

func editEvent(arg string) {
	log.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()

	req := eventv1.UpdateEventRequest{
		EventSelector: util.ResolveEvent(arg),
	}
	if editName != "" {
		req.Name = editName
	}
	if editDescription != "" {
		req.Description = editDescription
	}
	if editKey != "" {
		req.Key = editKey
	}
	//nolint:nestif // by design
	if editMinSessionTime != "" || editMaxSessionTime != "" {
		req.ReplayInfo = &eventv1.ReplayInfo{}
		if editMinSessionTime != "" {
			if v, err := time.ParseDuration(editMinSessionTime); err == nil {
				req.ReplayInfo.MinSessionTime = float32(v.Seconds())
			}
		}
		if editMaxSessionTime != "" {
			if v, err := time.ParseDuration(editMaxSessionTime); err == nil {
				req.ReplayInfo.MaxSessionTime = float32(v.Seconds())
			}
		}
	}
	md := metadata.Pairs("api-token", config.DefaultCliArgs().Token)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	c := eventv1grpc.NewEventServiceClient(conn)
	if _, err := c.UpdateEvent(ctx, &req); err != nil {
		log.Error("could not update event", log.ErrorField(err), log.String("event", arg))
		return
	}
	log.Info("Event updated.")
}
