package event

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/event/list"
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
	return cmd
}
