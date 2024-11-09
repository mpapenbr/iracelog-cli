package load

import (
	"context"

	eventv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
	"github.com/mpapenbr/iracelog-cli/util/output"
	"github.com/mpapenbr/iracelog-cli/util/output/event"
)

var (
	carNumFilter []string
	components   []string
	format       string
)

func NewEventLoadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "load stored event.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			loadEvent(args[0])
		},
	}
	cmd.Flags().StringSliceVar(&carNumFilter, "filter-carnum", []string{},
		"filter cars by car number")
	cmd.Flags().StringSliceVar(&components, "components", []string{"info", "cars"},
		"components to display (info, cars, carlaps)")
	cmd.Flags().StringVar(&format, "format", "text",
		"output format (text, json,csv)")

	return cmd
}

func loadEvent(arg string) {
	log.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()

	req := eventv1.GetEventRequest{
		EventSelector: util.ResolveEvent(arg),
	}

	ctx := context.Background()
	c := eventv1grpc.NewEventServiceClient(conn)
	var resp *eventv1.GetEventResponse
	if resp, err = c.GetEvent(ctx, &req); err != nil {
		log.Error("could not load event", log.ErrorField(err), log.String("event", arg))
		return
	}
	log.Info("Event loaded.")
	opts := []event.Option{}
	if len(carNumFilter) > 0 {
		opts = append(opts, event.WithCarNumFilter(carNumFilter))
	}
	if format != "" {
		if f, err := output.ParseFormat(format); err == nil {
			opts = append(opts, event.WithFormat(f))
		}
	}
	if len(components) > 0 {
		comps := []event.Component{}
		for _, c := range components {
			v, _ := event.ParseComponent(c)
			comps = append(comps, v)
		}
		opts = append(opts, event.WithComponents(comps))
	}

	out := event.NewEventOutput(resp, opts...)
	out.Output()
}
