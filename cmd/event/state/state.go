package state

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/event/state/car"
	"github.com/mpapenbr/iracelog-cli/cmd/event/state/options"
)

// eventCmd represents the event command

func NewStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "Commands for event state messages.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(car.NewStateCarCmd())
	cmd.PersistentFlags().DurationVar(&options.SessionTime, "session-time", 0,
		"session time as duration where data should begin (for example: 10m)")
	cmd.PersistentFlags().StringVar(&options.RecordStamp, "record-stamp", "",
		"timestamp time where data should begin")

	cmd.PersistentFlags().Int32Var(&options.NumEntries, "num", 0,
		"how many entries to check (0 means all)")

	cmd.MarkFlagsMutuallyExclusive("session-time", "record-stamp")
	cmd.MarkFlagsOneRequired("session-time", "record-stamp")
	return cmd
}
