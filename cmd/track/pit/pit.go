package pit

import (
	"github.com/spf13/cobra"
)

func NewPitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pit",
		Short: "Commands regarding pits. ",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.AddCommand(NewPitEditCmd())
	cmd.AddCommand(NewPitTransferCmd())

	return cmd
}
