package dummy

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/cmd/stress/config"
	"github.com/mpapenbr/iracelog-cli/log"
	myStress "github.com/mpapenbr/iracelog-cli/stress"
)

func NewStressDummyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dummy",
		Short: "dummy to test logger settings",
		Run: func(cmd *cobra.Command, args []string) {
			experimental(cmd.Context())
		},
	}

	return cmd
}

func experimental(ctx context.Context) {
	logger := log.GetFromContext(ctx)
	logger.Info("starting stress test")
	configOptions := config.CollectStandardJobProcessorOptions()
	configOptions = append(configOptions,

		myStress.WithLogging(logger),
		//nolint:gosec // ok here
		myStress.WithJobHandler(func(j *myStress.Job) error {
			j.Logger.Debug("about to sleep", log.Int("jobId", j.Id))
			waitTime := 100 + rand.Intn(100)
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
			j.Logger.Debug("done sleeping",
				log.Int("jobId", j.Id),
				log.Time("myTime", time.Now()))
			if rand.Intn(5) == 0 {
				return errors.New("simulated error")
			}
			return nil
		}))
	jobProcessor := myStress.NewJobProcessor(configOptions...)
	jobProcessor.Run()
}
