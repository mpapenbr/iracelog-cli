package live

import (
	"context"
	"math/rand"
	"time"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/provider/v1/providerv1grpc"
	providerv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/provider/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/mpapenbr/iracelog-cli/cmd/stress/config"
	appCfg "github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	myStress "github.com/mpapenbr/iracelog-cli/stress"
	"github.com/mpapenbr/iracelog-cli/util"
	"github.com/mpapenbr/iracelog-cli/util/simulate"
)

var (
	jobDurationArg   string
	jobDurationFixed bool
)

func NewStressLiveWebclientListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webclient",
		Short: "dummy to test stress component ",
		Run: func(cmd *cobra.Command, args []string) {
			webclient()
		},
	}
	cmd.Flags().StringVar(&jobDurationArg,
		"job-duration", "", "how long should a job run (example: \"10s\")")
	cmd.Flags().BoolVar(&jobDurationFixed,
		"job-duration-fixed",
		false,
		"if set, job duration is fixed, otherwise random up to job-duration")
	return cmd
}

// the plan for a job
// - get list of providers
// - pick one by random
// - be listener for live data for random time
// - done
//
//nolint:funlen,gocognit // ok here
func webclient() {
	logger := log.GetLoggerManager().GetDefaultLogger()
	configOptions := config.CollectStandardJobProcessorOptions()
	configOptions = append(configOptions,
		myStress.WithClientProvider(func() *grpc.ClientConn {
			c, err := util.ConnectGrpc(appCfg.DefaultCliArgs())
			if err != nil {
				logger.Fatal("could  not connect server", log.ErrorField(err))
			}
			logger.Debug("connected to server")
			return c
		}),
		myStress.WithJobHandler(func(j *myStress.Job) error {
			req := providerv1.ListLiveEventsRequest{}
			c := providerv1grpc.NewProviderServiceClient(j.Client)
			r, err := c.ListLiveEvents(context.Background(), &req)
			if err != nil {
				logger.Error("could not get live events", log.ErrorField(err))
				return err
			}
			if len(r.Events) == 0 {
				logger.Info("no events found")
				return nil
			}
			//nolint:gosec // ok here
			idx := rand.Intn(len(r.Events))
			logger.Info("picked event", log.Uint32("id", r.Events[idx].Event.Id))

			opts := []simulate.Option{
				simulate.WithClient(j.Client),
			}

			var ctx context.Context
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(j.Ctx)

			if jobDurationArg != "" {
				if d, err := time.ParseDuration(jobDurationArg); err == nil {
					if !jobDurationFixed {
						//nolint:gosec // ok here
						d = time.Duration((1 + rand.Intn(int(d.Seconds()))) * int(time.Second))
					}
					ctx, cancel = context.WithTimeout(j.Ctx, d)
					deadLine, _ := ctx.Deadline()
					logger.Info("job param",
						log.Duration("duration", d),
						log.Time("deadline", deadLine))
				}
			}
			defer cancel()

			opts = append(opts, simulate.WithContext(ctx))
			if config.WorkerProgressArg != "" {
				if d, err := time.ParseDuration(config.WorkerProgressArg); err == nil {
					opts = append(opts, simulate.WithStatsCallback(d, func(s *simulate.Stats) {
						logger.Info("stats", log.Any("stats", s))
					}))
				}
			}

			wc := simulate.NewWebclient(opts...)
			sel := util.ResolveEvent(r.Events[idx].Event.Key)
			//nolint:errcheck // by design
			wc.Start(sel)

			return nil
		}),
	)
	start := time.Now()
	jobProcessor := myStress.NewJobProcessor(configOptions...)
	jobProcessor.Run()
	logger.Info("job processor finished", log.Duration("duration", time.Since(start)))
}
