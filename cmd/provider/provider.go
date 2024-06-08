package provider

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/provider/list"
	"github.com/mpapenbr/iracelog-cli/cmd/provider/register"
	"github.com/mpapenbr/iracelog-cli/cmd/provider/unregister"
)

func NewProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Commands regarding data provider.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(list.NewProviderListCmd())
	cmd.AddCommand(register.NewProviderRegisterCmd())
	cmd.AddCommand(unregister.NewProviderUnregisterCmd())
	cmd.AddCommand(unregister.NewProviderUnregisterAllCmd())
	return cmd
}
