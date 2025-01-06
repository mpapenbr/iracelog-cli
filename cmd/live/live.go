package live

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/live/analysis"
	"github.com/mpapenbr/iracelog-cli/cmd/live/driver"
	"github.com/mpapenbr/iracelog-cli/cmd/live/snapshot"
	"github.com/mpapenbr/iracelog-cli/cmd/live/speedmap"
	"github.com/mpapenbr/iracelog-cli/cmd/live/state"
	"github.com/mpapenbr/iracelog-cli/cmd/live/webclient"
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
	cmd.AddCommand(driver.NewLiveDriverCmd())
	cmd.AddCommand(driver.NewSendEmptyDriverData())
	cmd.AddCommand(analysis.NewLiveAnalysisCmd())
	cmd.AddCommand(analysis.NewLiveAnalysisSelectorCmd())
	cmd.AddCommand(analysis.NewLiveCarOccupancyCmd())
	cmd.AddCommand(speedmap.NewLiveSpeedmapCmd())
	cmd.AddCommand(snapshot.NewLiveSnapshotCmd())
	cmd.AddCommand(webclient.NewLiveWebclientCmd())

	return cmd
}
