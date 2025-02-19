package list

import (
	"context"
	"time"

	tenantv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/tenant/v1/tenantv1grpc"
	tenantv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/tenant/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/iracelog-cli/cmd/tenant/cmdopts"
	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

var (
	attrs  []string
	format string
)

func NewTenantListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "lists tenants.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			listTenants(cmd.Context())
		},
	}
	cmd.Flags().StringSliceVar(&attrs, "attrs", []string{},
		"tenant attributes to display")
	cmd.Flags().StringVar(&format, "format", "text",
		"output format (text, json,csv)")
	return cmd
}

func listTenants(ctx context.Context) {
	logger := log.GetFromContext(ctx)
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	req := tenantv1.GetTenantsRequest{}

	c := tenantv1grpc.NewTenantServiceClient(conn)
	reqCtx, cancel := context.WithTimeout(
		metadata.NewOutgoingContext(context.Background(),
			metadata.Pairs(config.API_TOKEN_HEADER, config.DefaultCliArgs().Token)),
		10*time.Second)
	defer cancel()
	r, err := c.GetTenants(reqCtx, &req)
	if err != nil {
		logger.Error("could not get tenants", log.ErrorField(err))
		return
	}
	out := cmdopts.ConfigureOutput(format, attrs)

	out.Header()
	for _, t := range r.Tenants {
		out.Line(t)
	}
	out.Flush()
}
