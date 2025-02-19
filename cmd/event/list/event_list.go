package list

import (
	"context"
	"errors"
	"io"
	"time"

	eventv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewEventListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists stored events.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			listEvents(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&externalId,
		"tenant-external-id",
		"",
		"external id of the tenant")
	cmd.Flags().StringVar(&name,
		"tenant-name",
		"",
		"name of the tenant")
	return cmd
}

var (
	externalId string
	name       string
)

type (
	tenantParam struct{}
)

func (t tenantParam) ExternalId() string {
	return externalId
}

func (t tenantParam) Name() string {
	return name
}

func listEvents(ctx context.Context) {
	logger := log.GetFromContext(ctx)
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	req := eventv1.GetEventsRequest{
		TenantSelector: util.ResolveTenant(tenantParam{}),
	}
	c := eventv1grpc.NewEventServiceClient(conn)
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	r, err := c.GetEvents(reqCtx, &req)
	if err != nil {
		logger.Error("could not get events", log.ErrorField(err))
		return
	}

	for {
		resp, err := r.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			logger.Error("error fetching events", log.ErrorField(err))
			break
		} else {
			logger.Info("got event: ",
				log.Uint32("id", resp.Event.Id),
				log.String("key", resp.Event.Key))
		}
	}
}
