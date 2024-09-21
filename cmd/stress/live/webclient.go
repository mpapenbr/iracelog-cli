package live

import (
	"context"
	"math/rand"
	"sync"
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
	singleConnection bool
)

func NewStressLiveWebclientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webclient",
		Short: "simulate a set of webclients listening to live data",
		Run: func(cmd *cobra.Command, args []string) {
			webclient(cmd.Context())
		},
	}
	cmd.Flags().StringVar(&jobDurationArg,
		"job-duration", "", "how long should a job run (example: \"10s\")")
	cmd.Flags().BoolVar(&jobDurationFixed,
		"job-duration-fixed",
		false,
		"if set, job duration is fixed, otherwise random up to job-duration")
	cmd.Flags().BoolVar(&singleConnection,
		"single-connection",
		false,
		"if set, all threads will use the same server connection")
	return cmd
}

// the plan for a job
// - get list of providers
// - pick one by random
// - be listener for live data for random time
// - done
//
//nolint:funlen,gocognit,cyclop,gocyclo // ok here
func webclient(ctx context.Context) {
	logger := log.GetFromContext(ctx)
	statsLogger := logger.Named("stats")
	configOptions := config.CollectStandardJobProcessorOptions()
	var singleConn *grpc.ClientConn
	if singleConnection {
		var err error
		singleConn, err = util.ConnectGrpc(appCfg.DefaultCliArgs())
		if err != nil {
			logger.Fatal("could  not connect server", log.ErrorField(err))
		}
		logger.Debug("connected to server")
	}
	summary := newWebclientStats(config.WorkerThreads, logger.Named("summary"))
	configOptions = append(configOptions,
		myStress.WithLogging(logger),
		myStress.WithFinishHandler(func() {
			logger.Info("finish handler called")
			// ctx.Done()
		}),
		myStress.WithWorkerProgress(30*time.Second),
		myStress.WithTargetClientProvider(func() *grpc.ClientConn {
			if singleConnection {
				return singleConn
			} else {
				c, err := util.ConnectGrpc(appCfg.DefaultCliArgs())
				if err != nil {
					logger.Fatal("could  not connect server", log.ErrorField(err))
				}
				logger.Debug("connected to server")
				return c
			}
		}),
		myStress.WithJobHandler(func(j *myStress.Job) error {
			d, ok := j.Ctx.Deadline()
			j.Logger.Info("deadline", log.Time("deadline", d), log.Bool("ok", ok))
			req := providerv1.ListLiveEventsRequest{}
			c := providerv1grpc.NewProviderServiceClient(j.TargetClient)
			r, err := c.ListLiveEvents(j.Ctx, &req)
			if err != nil {
				j.Logger.Error("could not get live events", log.ErrorField(err))
				return err
			}
			if len(r.Events) == 0 {
				j.Logger.Info("no events found")
				time.Sleep(1 * time.Second)
				return nil
			}
			//nolint:gosec // ok here
			idx := rand.Intn(len(r.Events))
			j.Logger.Info("picked event", log.Uint32("id", r.Events[idx].Event.Id))

			opts := []simulate.Option{
				simulate.WithClient(j.TargetClient),
			}

			var jobCtx context.Context
			var cancel context.CancelFunc
			jobCtx, cancel = context.WithCancel(j.Ctx)

			if jobDurationArg != "" {
				if d, err := time.ParseDuration(jobDurationArg); err == nil {
					if !jobDurationFixed {
						//nolint:gosec // ok here
						d = time.Duration((1 + rand.Intn(int(d.Seconds()))) * int(time.Second))
					}
					jobCtx, cancel = context.WithTimeout(j.Ctx, d)
					deadLine, _ := jobCtx.Deadline()
					j.Logger.Info("job param",
						log.Duration("duration", d),
						log.Time("deadline", deadLine))
				}
			}
			defer cancel()

			opts = append(opts, simulate.WithContext(jobCtx))
			if config.WorkerProgressArg != "" {
				if d, err := time.ParseDuration(config.WorkerProgressArg); err == nil {
					opts = append(opts, simulate.WithStatsCallback(d, func(s *simulate.Stats) {
						j.Logger.Info("stats", log.Any("stats", s))
					}))
				}
			}

			wc := simulate.NewWebclient(opts...)
			sel := util.ResolveEvent(r.Events[idx].Event.Key)

			if wcErr := wc.Start(sel); wcErr == nil {
				stats := wc.GetStats()
				statsLogger.Info("webclient finished",
					log.Int("jobId", j.Id),
					log.Int("workerId", j.WorkerId),
					log.Any("stats", stats))
				summary.addStats(j.WorkerId, &stats)
			} else {
				j.Logger.Error("webclient failed", log.ErrorField(wcErr))
			}

			return nil
		}),
	)
	start := time.Now()

	ticker := time.NewTicker(10 * time.Hour)
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("testDuration reached, terminating workerProgress")
				ticker.Stop()
				return
			case <-ticker.C:
				logger.Info("About to show progress of workers")
				summary.output()
			}
		}
	}()

	jobProcessor := myStress.NewJobProcessor(configOptions...)
	jobProcessor.Run()
	logger.Info("job processor finished", log.Duration("duration", time.Since(start)))
	summary.output()
}

type webclientStats struct {
	stats  []simulate.Stats
	mu     sync.Mutex
	logger *log.Logger
}

func newWebclientStats(numWorker int, logger *log.Logger) *webclientStats {
	return &webclientStats{
		stats:  make([]simulate.Stats, numWorker),
		mu:     sync.Mutex{},
		logger: logger,
	}
}

func (w *webclientStats) output() {
	for i := range w.stats {
		w.logger.Info("summary", log.Int("workerId", i), log.Any("stats", w.stats[i]))
	}
}

func (w *webclientStats) addStats(workerId int, s *simulate.Stats) {
	w.mu.Lock()
	defer w.mu.Unlock()
	item := &w.stats[workerId]
	item.Analysis.Add(&s.Analysis)
	item.Driver.Add(&s.Driver)
	item.Speedmap.Add(&s.Speedmap)
	item.State.Add(&s.State)
}
