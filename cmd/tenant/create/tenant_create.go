package create

import (
	"context"
	"fmt"
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
	apiKey    string
	apiKeyLen uint
	attrs     []string
	format    string
	active    bool
)

func NewTenantCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create tenant.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			createTenant(cmd.Context(), args[0])
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&apiKey,
		"api-key", "", "assign api key to tenant")
	cmd.Flags().UintVar(&apiKeyLen,
		"api-key-len", 32, "length of generated api key")
	cmd.Flags().BoolVar(&active,
		"active", true, "active state of tenant")
	cmd.Flags().StringSliceVar(&attrs, "attrs", []string{},
		"tenant attributes to display")
	cmd.Flags().StringVar(&format, "format", "text",
		"output format (text, json,csv)")

	return cmd
}

func createTenant(ctx context.Context, name string) {
	logger := log.GetFromContext(ctx)
	logger.Debug("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	if apiKey == "" {
		if apiKey, err = util.GenerateRandomString(apiKeyLen); err != nil {
			logger.Fatal("could not generate random string", log.ErrorField(err))
		}
	}
	req := tenantv1.CreateTenantRequest{
		Name:     name,
		IsActive: active,
		ApiKey:   apiKey,
	}
	c := tenantv1grpc.NewTenantServiceClient(conn)
	reqCtx, cancel := context.WithTimeout(
		metadata.NewOutgoingContext(context.Background(),
			metadata.Pairs(config.API_TOKEN_HEADER, config.DefaultCliArgs().Token)),
		10*time.Second)
	defer cancel()
	r, err := c.CreateTenant(reqCtx, &req)
	if err != nil {
		logger.Error("could not create tenant", log.ErrorField(err))
		return
	}
	fmt.Println("Tenant created")
	fmt.Println("The api key is: ", apiKey)
	out := cmdopts.ConfigureOutput(format, attrs)
	out.Header()
	out.Line(r.Tenant)
	out.Flush()
}
