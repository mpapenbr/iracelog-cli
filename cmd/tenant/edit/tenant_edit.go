package edit

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
	name           string
	newname        string
	apiKey         string
	apiKeyLen      uint
	generateApiKey bool
	attrs          []string
	format         string
	enableActive   bool
	disableActive  bool
	externalId     string
)

func NewTenantEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "edit tenant data.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			editTenant(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&newname,
		"new-name", "", "assign new name to tenant")
	cmd.Flags().StringVar(&apiKey,
		"api-key", "", "assign api key to tenant")
	cmd.Flags().UintVar(&apiKeyLen,
		"api-key-len", 32, "length of generated api key")
	cmd.Flags().BoolVar(&generateApiKey,
		"generate-api-key", false, "generate a new api key")
	cmd.Flags().BoolVar(&enableActive,
		"enable-active", false, "set active state of tenant to true")
	cmd.Flags().BoolVar(&disableActive,
		"disable-active", false, "set active state of tenant to false")
	cmd.Flags().StringSliceVar(&attrs, "attrs", []string{},
		"tenant attributes to display")
	cmd.Flags().StringVar(&format, "format", "text",
		"output format (text, json,csv)")
	cmd.Flags().StringVar(&externalId,
		"external-id",
		"",
		"external id of the tenant")
	cmd.Flags().StringVar(&name,
		"name",
		"",
		"name of the tenant")
	cmd.MarkFlagsMutuallyExclusive("enable-active", "disable-active")
	cmd.MarkFlagsMutuallyExclusive("api-key", "generate-api-key")
	cmd.MarkFlagsOneRequired("name", "external-id")
	return cmd
}

type (
	tenantParam struct{}
)

func (t tenantParam) ExternalId() string {
	return externalId
}

func (t tenantParam) Name() string {
	return name
}

//nolint:funlen // by design
func editTenant(ctx context.Context) {
	logger := log.GetFromContext(ctx)
	logger.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		logger.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()
	sel := util.ResolveTenant(tenantParam{})
	c := tenantv1grpc.NewTenantServiceClient(conn)
	reqCtx, cancel := context.WithTimeout(
		metadata.NewOutgoingContext(context.Background(),
			metadata.Pairs(config.API_TOKEN_HEADER, config.DefaultCliArgs().Token)),
		10*time.Second)
	defer cancel()

	// request tenant data
	var cur *tenantv1.GetTenantResponse
	if cur, err = c.GetTenant(reqCtx,
		&tenantv1.GetTenantRequest{Tenant: sel}); err != nil {
		logger.Fatal("could not get tenant", log.ErrorField(err))
	}

	req := tenantv1.UpdateTenantRequest{
		Tenant:   sel,
		IsActive: cur.Tenant.IsActive,
	}
	if newname != "" {
		req.Name = newname
	}
	if generateApiKey {
		if apiKey, err = util.GenerateRandomString(apiKeyLen); err != nil {
			logger.Fatal("could not generate random string", log.ErrorField(err))
		}
		req.ApiKey = apiKey
	}
	if apiKey != "" {
		req.ApiKey = apiKey
	}
	if enableActive {
		req.IsActive = true
	}
	if disableActive {
		req.IsActive = false
	}

	r, err := c.UpdateTenant(reqCtx, &req)
	if err != nil {
		logger.Error("could not get tenants", log.ErrorField(err))
		return
	}

	fmt.Println("Tenant updated")
	out := cmdopts.ConfigureOutput(format, attrs)
	out.Header()
	out.Line(r.Tenant)
	out.Flush()
}
