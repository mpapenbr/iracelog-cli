/*
 */
package stress

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/stress/config"
	"github.com/mpapenbr/iracelog-cli/cmd/stress/dummy"
	"github.com/mpapenbr/iracelog-cli/cmd/stress/events"
	"github.com/mpapenbr/iracelog-cli/cmd/stress/live"
	"github.com/mpapenbr/iracelog-cli/cmd/stress/replay"
)

var (
	testDurationArg = "10m"
	pauseArg        = "1s"
)

func NewStressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stress",
		Short: "stress mvk-repos with actions around dvl functions",
	}
	cmd.PersistentFlags().IntVarP(&config.WorkerThreads,
		"worker", "w", 1, "Amount of worker threads")
	cmd.PersistentFlags().StringVarP(&config.TestDuration,
		"duration", "d", testDurationArg, "Duration of stress test")
	cmd.PersistentFlags().StringVar(&config.Pause,
		"pause", pauseArg,
		"max. pause before next iteration is issued (will use random value)")

	cmd.PersistentFlags().StringVar(&config.WorkerProgressArg,
		"worker-stats", "", "interval for showing worker progress stats (example: \"10s\")")

	cmd.PersistentFlags().StringVar(&config.JobLogLevelArg,
		"job-log-level", "info", "log level for job processing component")

	cmd.AddCommand(dummy.NewStressDummyCmd())
	cmd.AddCommand(events.NewStressDummyEventListCmd())
	cmd.AddCommand(live.NewStressLiveWebclientCmd())
	cmd.AddCommand(replay.NewStressReplayCmd())
	return cmd
}
