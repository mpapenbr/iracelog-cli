package deleteit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	tenantv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/tenant/v1/tenantv1grpc"
	tenantv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/tenant/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewTenantDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete tenant.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			if id, err := strconv.Atoi(args[0]); err == nil {
				deleteTenant(cmd.Context(), uint32(id))
			}
		},
		Args: cobra.ExactArgs(1),
	}
	return cmd
}

func deleteTenant(ctx context.Context, id uint32) {
	logger := log.GetFromContext(ctx)
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	req := tenantv1.DeleteTenantRequest{
		Id: id,
	}
	c := tenantv1grpc.NewTenantServiceClient(conn)
	reqCtx, cancel := context.WithTimeout(
		metadata.NewOutgoingContext(context.Background(),
			metadata.Pairs(config.API_TOKEN_HEADER, config.DefaultCliArgs().Token)),
		10*time.Second)
	defer cancel()
	_, err = c.DeleteTenant(reqCtx, &req)
	if err != nil {
		logger.Error("could not delete tenant", log.ErrorField(err))
		return
	}

	fmt.Println("Tenant deleted")
}
