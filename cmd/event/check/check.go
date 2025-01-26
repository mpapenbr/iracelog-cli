package check

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/event/check/driver"
	"github.com/mpapenbr/iracelog-cli/cmd/event/check/options"
	"github.com/mpapenbr/iracelog-cli/cmd/event/check/speedmap"
	"github.com/mpapenbr/iracelog-cli/cmd/event/check/state"
	"github.com/mpapenbr/iracelog-cli/cmd/event/check/tire"
)

func NewCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Commands to check data consistency.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(state.NewCheckStateCmd())

	cmd.AddCommand(speedmap.NewCheckSpeedmapCmd())
	cmd.AddCommand(driver.NewCheckDriverCmd())
	cmd.AddCommand(tire.NewCheckTireCmd())

	cmd.PersistentFlags().DurationVar(&options.SessionTime, "session-time", 0,
		"session time as duration where data should begin (for example: 10m)")
	cmd.PersistentFlags().StringVar(&options.RecordStamp, "record-stamp", "",
		"timestamp time where data should begin")
	cmd.PersistentFlags().DurationVar(&options.GapThreshold, "gap", 2*time.Second,
		"gap threshold for warning about gaps in data")
	cmd.PersistentFlags().Int32Var(&options.NumEntries, "num", 0,
		"how many entries to check (0 means all)")

	cmd.MarkFlagsMutuallyExclusive("session-time", "record-stamp")
	cmd.MarkFlagsOneRequired("session-time", "record-stamp")
	return cmd
}
