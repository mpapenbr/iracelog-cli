/*
 */
package stress

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/stress/config"
	"github.com/mpapenbr/iracelog-cli/cmd/stress/dummy"
	"github.com/mpapenbr/iracelog-cli/cmd/stress/events"
	"github.com/mpapenbr/iracelog-cli/cmd/stress/live"
	"github.com/mpapenbr/iracelog-cli/cmd/stress/replay"
)

func NewStressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stress",
		Short: "stress mvk-repos with actions around dvl functions",
	}
	cmd.PersistentFlags().IntVarP(&config.WorkerThreads,
		"worker", "w", 1, "Amount of worker threads")

	cmd.PersistentFlags().DurationVar(&config.RampUpDuration,
		"rampup-duration", 0, "how long should a ramp up phase last (example \"5m\")")
	cmd.PersistentFlags().IntVar(&config.RampUpIncrease,
		"rampup-increase", 0, "Increase worker amount during ramp up")
	cmd.PersistentFlags().IntVar(&config.RampUpInitial,
		"rampup-initial", 1, "Initial amount of workers during ramp up")

	cmd.PersistentFlags().DurationVarP(&config.TestDuration,
		"duration", "d", 10*time.Minute, "Duration of stress test")
	cmd.PersistentFlags().DurationVar(&config.Pause,
		"pause", 2*time.Second,
		"max. pause before next iteration is issued (will use random value)")

	cmd.PersistentFlags().DurationVar(&config.WorkerProgress,
		"worker-stats", 0, "interval for showing worker progress stats (example: \"10s\")")

	cmd.PersistentFlags().StringVar(&config.JobLogLevelArg,
		"job-log-level", "info", "log level for job processing component")

	cmd.AddCommand(dummy.NewStressDummyCmd())
	cmd.AddCommand(events.NewStressDummyEventListCmd())
	cmd.AddCommand(live.NewStressLiveWebclientCmd())
	cmd.AddCommand(replay.NewStressReplayCmd())
	return cmd
}
