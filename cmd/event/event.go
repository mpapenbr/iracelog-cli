package event

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/event/analysis"
	"github.com/mpapenbr/iracelog-cli/cmd/event/check"
	"github.com/mpapenbr/iracelog-cli/cmd/event/deleteit"
	"github.com/mpapenbr/iracelog-cli/cmd/event/edit"
	"github.com/mpapenbr/iracelog-cli/cmd/event/list"
	"github.com/mpapenbr/iracelog-cli/cmd/event/load"
	"github.com/mpapenbr/iracelog-cli/cmd/event/replay"
	"github.com/mpapenbr/iracelog-cli/cmd/event/session"
	"github.com/mpapenbr/iracelog-cli/cmd/event/state"
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
	cmd.AddCommand(edit.NewEventEditCmd())
	cmd.AddCommand(replay.NewEventReplayCmd())
	cmd.AddCommand(session.NewEventSessionCmd())
	cmd.AddCommand(check.NewCheckCmd())
	cmd.AddCommand(state.NewStateCmd())
	cmd.AddCommand(analysis.NewEventAnalysisComputeCmd())
	return cmd
}
