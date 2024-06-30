package dummy

import (
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
		Short: "dummy to test stress component ",
		Run: func(cmd *cobra.Command, args []string) {
			experimental()
		},
	}

	return cmd
}

func experimental() {
	logger := log.GetLoggerManager().GetDefaultLogger()
	configOptions := config.CollectStandardJobProcessorOptions()
	configOptions = append(configOptions,
		//nolint:gosec // by design
		myStress.WithJobHandler(func(j *myStress.Job) error {
			logger.Debug("about to sleep", log.Int("jobId", j.Id))
			waitTime := 100 + rand.Intn(100)
			time.Sleep(time.Duration(waitTime) * time.Millisecond)
			logger.Debug("done sleeping", log.Int("jobId", j.Id), log.Time("myTime", time.Now()))
			if rand.Intn(5) == 0 {
				return errors.New("simulated error")
			}
			return nil
		}))
	jobProcessor := myStress.NewJobProcessor(configOptions...)
	jobProcessor.Run()
}
