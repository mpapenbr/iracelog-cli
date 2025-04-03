package replayloop

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"slices"
	"sync"
	"time"

	eventv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	providerv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/provider/v1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	myStress "github.com/mpapenbr/iracelog-cli/stress"
	"github.com/mpapenbr/iracelog-cli/util"
	"github.com/mpapenbr/iracelog-cli/util/replay"
)

var (
	jobDuration      time.Duration
	jobDurationFixed bool
	demoPrefix       string
	cfg              *replay.Config
	includeEvents    []uint
	excludeEvents    []uint
	useJobIDKey      bool
)

func init() {
	cfg = replay.DefaultConfig()
}

//nolint:funlen // by design
func NewReplayLoopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replayloop",
		Short: "replay events in an inifinite loop.",

		Run: func(cmd *cobra.Command, args []string) {
			replayLoop(cmd.Context())
		},
	}
	cmd.Flags().IntVarP(&myStress.WorkerThreads,
		"worker", "w", 1, "Amount of worker threads")
	cmd.Flags().DurationVarP(&myStress.TestDuration,
		"duration", "d", 0*time.Minute, "Duration of stress test")
	cmd.Flags().UintSliceVar(&includeEvents,
		"include-events", []uint{}, "include events by id")
	cmd.Flags().UintSliceVar(&excludeEvents,
		"exclude-events", []uint{}, "exclude events by id")
	cmd.Flags().BoolVar(&useJobIDKey,
		"use-jobid-key", true,
		"use job id for event key")
	cmd.Flags().DurationVar(&myStress.Pause,
		"pause", 2*time.Second,
		"max. pause before next iteration is issued (will use random value)")

	cmd.Flags().IntVar(&cfg.Speed, "speed", 1,
		"Recording speed (0 means: go as fast as possible)")
	cmd.Flags().StringVar(&cfg.SourceAddr,
		"source-addr",
		"",
		"gRPC server address")
	cmd.Flags().BoolVar(&cfg.SourceInsecure,
		"source-insecure",
		false,
		"connect gRPC address without TLS (development only)")

	cmd.Flags().StringVarP(&cfg.Token,
		"token", "t", "", "authentication token")

	cmd.Flags().BoolVar(&cfg.DoNotPersist,
		"do-not-persist", false, "do not persist data")
	cmd.Flags().DurationVar(&cfg.FastForward,
		"fast-forward",
		time.Duration(0),
		"replay this duration with max speed (relative to first event timestamp)")
	cmd.Flags().BoolVar(&cfg.FFPreRace,
		"ff-prerace", true, "fast forward prerace events")
	cmd.Flags().StringVar(&demoPrefix, "demo-prefix",
		"Demo replay of", "prefix for demo events")
	cmd.MarkFlagsMutuallyExclusive("include-events", "exclude-events")
	return cmd
}

//nolint:funlen,gocognit,cyclop,gocyclo //by design
func replayLoop(ctx context.Context) {
	logger := log.GetFromContext(ctx)
	mutex := sync.Mutex{}
	activeReplayEvents := []uint{}

	myCtx, cancel := context.WithCancel(ctx)
	configOptions := myStress.CollectStandardJobProcessorOptions()
	configOptions = append(configOptions,
		myStress.WithLogging(logger),
		myStress.WithContext(myCtx),
		myStress.WithSourceClientProvider(func() *grpc.ClientConn {
			c, err := util.NewClient(cfg.SourceAddr, util.WithTLSEnabled(!cfg.SourceInsecure))
			if err != nil {
				logger.Fatal("could not connect source server", log.ErrorField(err))
			}
			logger.Debug("connected to source server")
			return c
		}),
		myStress.WithTargetClientProvider(func() *grpc.ClientConn {
			c, err := util.ConnectGrpc(config.DefaultCliArgs())
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
			workEvents := r.Events
			// clear execluded excluded events
			if len(excludeEvents) > 0 {
				workEvents = slices.DeleteFunc(r.Events, func(check *eventv1.Event) bool {
					return slices.ContainsFunc(excludeEvents, func(item uint) bool {
						return item == uint(check.Id)
					})
				})
			}
			// only include these events
			if len(includeEvents) > 0 {
				workEvents = slices.DeleteFunc(workEvents, func(check *eventv1.Event) bool {
					return !slices.ContainsFunc(includeEvents, func(item uint) bool {
						return item == uint(check.Id)
					})
				})
			}
			mutex.Lock()
			// remove events that are already replayed
			workEvents = slices.DeleteFunc(workEvents, func(check *eventv1.Event) bool {
				return slices.ContainsFunc(activeReplayEvents, func(item uint) bool {
					return item == uint(check.Id)
				})
			})
			for _, e := range workEvents {
				j.Logger.Info("event", log.Uint32("id", e.Id), log.String("name", e.Name))
			}

			//nolint:gosec // ok here
			idx := rand.Intn(len(workEvents))
			j.Logger.Info("picked event", log.Uint32("id", workEvents[idx].Id))
			e := workEvents[idx]
			activeReplayEvents = append(activeReplayEvents, uint(e.Id))
			mutex.Unlock()

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
			opts := []replay.ReplayOption{}

			var ctx context.Context
			var jobCancel context.CancelFunc
			ctx, jobCancel = context.WithCancel(j.Ctx)

			if jobDuration > 0 {
				d := jobDuration
				if !jobDurationFixed {
					//nolint:gosec // ok here
					d = time.Duration((1 + rand.Intn(int(d.Seconds()))) * int(time.Second))
				}
				ctx, jobCancel = context.WithTimeout(j.Ctx, d)
				deadLine, _ := ctx.Deadline()
				j.Logger.Info("job param",
					log.Duration("duration", d),
					log.Time("deadline", deadLine))
			}
			defer jobCancel()

			opts = append(opts, replay.WithFastForward(cfg.FastForward))

			if cfg.Speed > 0 {
				opts = append(opts, replay.WithSpeed(cfg.Speed))
			}
			opts = append(opts, replay.WithContext(ctx))
			if cfg.Token != "" {
				opts = append(opts, replay.WithTokenProvider(func() string {
					return cfg.Token
				}))
			}

			dp := replay.NewDataProvider(j.SourceClient, e.Id,
				func() *providerv1.RegisterEventRequest {
					recordingMode := func() providerv1.RecordingMode {
						if cfg.DoNotPersist {
							return providerv1.RecordingMode_RECORDING_MODE_DO_NOT_PERSIST
						} else {
							return providerv1.RecordingMode_RECORDING_MODE_PERSIST
						}
					}
					//nolint:errcheck // ok here
					demoEvent := proto.Clone(e).(*eventv1.Event)
					if useJobIDKey {
						demoEvent.Key = fmt.Sprintf("demo-worker-%d", j.ID)
						e.Key = fmt.Sprintf("demo-worker-%d", j.ID)
					} else {
						demoEvent.Key = uuid.New().String()
					}
					demoEvent.Name = demoPrefix + " " + e.Name
					return &providerv1.RegisterEventRequest{
						Key:           demoEvent.Key,
						Event:         demoEvent,
						Track:         eventResp.Track,
						RecordingMode: recordingMode(),
					}
				})

			testMode := false
			if testMode {
				j.Logger.Debug("picked event", log.Uint32("id", e.Id), log.String("name", e.Name))
				j.Logger.Debug("test mode, not replaying event. Just sleeping")
				time.Sleep(5 * time.Second)
				j.Logger.Debug("Done")
			} else {
				rt := replay.NewReplayTask(j.TargetClient, dp, opts...)
				if err := rt.Replay(e.Id); err != nil {
					j.Logger.Error("error replaying event", log.ErrorField(err))
				}
			}
			mutex.Lock()
			activeReplayEvents = slices.DeleteFunc(activeReplayEvents,
				func(check uint) bool {
					return check == uint(e.Id)
				})
			mutex.Unlock()
			return nil
		}),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	defer func() {
		signal.Stop(sigChan)
		cancel()
	}()
	go func() {
		start := time.Now()
		jobProcessor := myStress.NewJobProcessor(configOptions...)
		jobProcessor.Run()
		logger.Info("job processor finished", log.Duration("duration", time.Since(start)))
		cancel()
	}()
	for {
		select {
		case <-sigChan:
			logger.Info("received interrupt signal, stopping")
			cancel()
			return
		case <-myCtx.Done():
			log.Debug("Received myCtx.Done")
			return
		}
	}
}
