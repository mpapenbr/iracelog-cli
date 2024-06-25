package event

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/event/deleteit"
	"github.com/mpapenbr/iracelog-cli/cmd/event/list"
	"github.com/mpapenbr/iracelog-cli/cmd/event/load"
)

// eventCmd represents the event command

func NewEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "Commands regarding events. ",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(list.NewEventListCmd())
	cmd.AddCommand(deleteit.NewEventDeleteCmd())
	cmd.AddCommand(load.NewEventLoadCmd())
	return cmd
}
