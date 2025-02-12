package state

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/event/state/car"
	"github.com/mpapenbr/iracelog-cli/cmd/event/state/driver"
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
	cmd.AddCommand(driver.NewStateDriverDataCmd())
	cmd.PersistentFlags().DurationVar(&options.SessionTime, "session-time", -1,
		"session time as duration where data should begin (for example: 10m)")
	cmd.PersistentFlags().Int32Var(&options.SessionNum, "session-num", -1,
		"session num to be used (-1 means: latest session)")
	cmd.PersistentFlags().StringVar(&options.RecordStamp, "record-stamp", "",
		"timestamp time where data should begin")
	cmd.PersistentFlags().Int32Var(&options.Id, "id", -1,
		"sequence id to be used for start selector (internal sequence)")

	cmd.PersistentFlags().Int32Var(&options.NumEntries, "num", 0,
		"number of entries to fetch (0 means all)")

	cmd.MarkFlagsMutuallyExclusive("session-time", "record-stamp", "id")
	cmd.MarkFlagsOneRequired("session-time", "record-stamp", "id")
	return cmd
}
