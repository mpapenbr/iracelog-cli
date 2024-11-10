package replay

import (
	"context"
	"math/rand"
	"time"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	providerv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/provider/v1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/mpapenbr/iracelog-cli/cmd/stress/config"
	appCfg "github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	myStress "github.com/mpapenbr/iracelog-cli/stress"
	"github.com/mpapenbr/iracelog-cli/util"
	utilReplay "github.com/mpapenbr/iracelog-cli/util/replay"
)

var (
	jobDuration      time.Duration
	jobDurationFixed bool
	cfg              *utilReplay.Config
)

func init() {
	cfg = utilReplay.DefaultConfig()
}

func NewStressReplayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replay",
		Short: "simulate a set of data providers by replaying events",
		Run: func(cmd *cobra.Command, args []string) {
			replay(cmd.Context())
		},
	}
	cmd.PersistentFlags().IntVar(&cfg.Speed, "speed", 1,
		"Recording speed (0 means: go as fast as possible)")
	cmd.PersistentFlags().StringVar(&cfg.SourceAddr,
		"source-addr",
		"",
		"gRPC server address")
	cmd.PersistentFlags().BoolVar(&cfg.SourceInsecure,
		"source-insecure",
		false,
		"connect gRPC address without TLS (development only)")

	cmd.PersistentFlags().StringVarP(&cfg.Token,
		"token", "t", "", "authentication token")
	cmd.Flags().DurationVar(&jobDuration,
		"job-duration", 0, "how long should a job run (example: \"10s\")")
	cmd.Flags().BoolVar(&jobDurationFixed,
		"job-duration-fixed",
		false,
		"if set, job duration is fixed, otherwise random up to job-duration")
	cmd.PersistentFlags().BoolVar(&cfg.DoNotPersist,
		"do-not-persist", false, "do not persist data")
	cmd.PersistentFlags().DurationVar(&cfg.FastForward,
		"fast-forward",
		time.Duration(0),
		"replay this duration with max speed")
	return cmd
}

// the plan for a job
// - get list of providers
// - pick one by random
// - be listener for live data for random time
// - done
//
//nolint:funlen,gocognit,cyclop // ok here
func replay(ctx context.Context) {
	logger := log.GetFromContext(ctx)
	configOptions := config.CollectStandardJobProcessorOptions()
	configOptions = append(configOptions,
		myStress.WithLogging(logger),
		myStress.WithSourceClientProvider(func() *grpc.ClientConn {
			c, err := util.NewClient(cfg.SourceAddr, util.WithTLSEnabled(!cfg.SourceInsecure))
			if err != nil {
				logger.Fatal("could not connect source server", log.ErrorField(err))
			}
			logger.Debug("connected to source server")
			return c
		}),
		myStress.WithTargetClientProvider(func() *grpc.ClientConn {
			c, err := util.ConnectGrpc(appCfg.DefaultCliArgs())
			if err != nil {
				logger.Fatal("could not connect target server", log.ErrorField(err))
			}
			logger.Debug("connected to target server")
			return c
		}),
		myStress.WithJobHandler(func(j *myStress.Job) error {
			req := eventv1.GetLatestEventsRequest{}
			c := eventv1grpc.NewEventServiceClient(j.SourceClient)
			r, err := c.GetLatestEvents(context.Background(), &req)
			if err != nil {
				j.Logger.Error("could not get events", log.ErrorField(err))
				return err
			}
			if len(r.Events) == 0 {
				j.Logger.Info("no events found")
				return nil
			}
			//nolint:gosec // ok here
			idx := rand.Intn(len(r.Events))
			j.Logger.Info("picked event", log.Uint32("id", r.Events[idx].Id))
			e := r.Events[idx]

			reqEvent := eventv1.GetEventRequest{
				EventSelector: &commonv1.EventSelector{Arg: &commonv1.EventSelector_Id{
					Id: int32(e.Id),
				}},
			}
			var eventResp *eventv1.GetEventResponse
			eventResp, err = c.GetEvent(context.Background(), &reqEvent)
			if err != nil {
				j.Logger.Error("could not get events", log.ErrorField(err))
				return err
			}
			opts := []utilReplay.ReplayOption{}

			var ctx context.Context
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(j.Ctx)

			if jobDuration > 0 {
				d := jobDuration
				if !jobDurationFixed {
					//nolint:gosec // ok here
					d = time.Duration((1 + rand.Intn(int(d.Seconds()))) * int(time.Second))
				}
				ctx, cancel = context.WithTimeout(j.Ctx, d)
				deadLine, _ := ctx.Deadline()
				j.Logger.Info("job param",
					log.Duration("duration", d),
					log.Time("deadline", deadLine))
			}
			defer cancel()

			opts = append(opts, utilReplay.WithFastForward(cfg.FastForward))

			if cfg.Speed > 0 {
				opts = append(opts, utilReplay.WithSpeed(cfg.Speed))
			}
			opts = append(opts, utilReplay.WithContext(ctx))
			if cfg.Token != "" {
				opts = append(opts, utilReplay.WithTokenProvider(func() string {
					return cfg.Token
				}))
			}
			//nolint:gocritic // by design
			// if config.WorkerProgressArg != "" {
			// 	if d, err := time.ParseDuration(config.WorkerProgressArg); err == nil {
			// 		opts = append(opts, utilReplay.WithStatsCallback(d, func(s *simulate.Stats) {
			// 			log.Info("stats", log.Any("stats", s))
			// 		}))
			// 	}
			// }

			dp := utilReplay.NewDataProvider(j.SourceClient, e.Id,
				func() *providerv1.RegisterEventRequest {
					recordingMode := func() providerv1.RecordingMode {
						if cfg.DoNotPersist {
							return providerv1.RecordingMode_RECORDING_MODE_DO_NOT_PERSIST
						} else {
							return providerv1.RecordingMode_RECORDING_MODE_PERSIST
						}
					}

					e.Key = uuid.New().String()
					return &providerv1.RegisterEventRequest{
						Event:         e,
						Track:         eventResp.Track,
						RecordingMode: recordingMode(),
					}
				})

			rt := utilReplay.NewReplayTask(j.TargetClient, dp, opts...)
			if err := rt.Replay(e.Id); err != nil {
				j.Logger.Error("error replaying event", log.ErrorField(err))
			}

			return nil
		}),
	)
	start := time.Now()
	jobProcessor := myStress.NewJobProcessor(configOptions...)
	jobProcessor.Run()
	logger.Info("job processor finished", log.Duration("duration", time.Since(start)))
}
