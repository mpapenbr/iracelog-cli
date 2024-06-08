package live

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/live/analysis"
	"github.com/mpapenbr/iracelog-cli/cmd/live/speedmap"
	"github.com/mpapenbr/iracelog-cli/cmd/live/state"
	"github.com/mpapenbr/iracelog-cli/config"
)

func NewLiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "live",
		Short: "Commands regarding live data.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&config.DefaultCliArgs().Event,
		"event", "", "event name")
	cmd.AddCommand(state.NewLiveStateCmd())
	cmd.AddCommand(analysis.NewLiveAnalysisCmd())
	cmd.AddCommand(analysis.NewLiveAnalysisSelectorCmd())
	cmd.AddCommand(analysis.NewLiveCarOccupancyCmd())
	cmd.AddCommand(speedmap.NewLiveSpeedmapCmd())

	return cmd
}
