package tenant

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/tenant/create"
	"github.com/mpapenbr/iracelog-cli/cmd/tenant/deleteit"
	"github.com/mpapenbr/iracelog-cli/cmd/tenant/edit"
	"github.com/mpapenbr/iracelog-cli/cmd/tenant/list"
	"github.com/mpapenbr/iracelog-cli/config"
)

func NewTenantCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Commands regarding tenants. ",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(list.NewTenantListCmd())
	cmd.AddCommand(create.NewTenantCreateCmd())
	cmd.AddCommand(deleteit.NewTenantDeleteCmd())
	cmd.AddCommand(edit.NewTenantEditCmd())

	cmd.PersistentFlags().StringVarP(&config.DefaultCliArgs().Token,
		"token", "t", "", "authentication token")
	return cmd
}
