package replay

import (
	"context"
	"time"

	eventv1grpc "buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/event/v1/eventv1grpc"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	providerv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/provider/v1"
	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
	"github.com/mpapenbr/iracelog-cli/util/replay"
)

var cfg *replay.Config

func init() {
	cfg = replay.DefaultConfig()
}

func NewEventReplayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replay",
		Short: "replay an event.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			replayEvent(args[0])
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
	cmd.PersistentFlags().StringVar(&cfg.EventKey,
		"key", "", "event key to use for replay")
	cmd.PersistentFlags().BoolVar(&cfg.DoNotPersist,
		"do-not-persist", false, "do not persist data")
	cmd.PersistentFlags().DurationVar(&cfg.FastForward,
		"fast-forward",
		time.Duration(0),
		"replay this duration with max speed")

	return cmd
}

//nolint:funlen,gocognit,cyclop //by design
func replayEvent(arg string) {
	log.Info("connect source server", log.String("addr", cfg.SourceAddr))
	source, err := util.NewClient(
		cfg.SourceAddr,
		util.WithTLSEnabled(!cfg.SourceInsecure))
	if err != nil {
		log.Error("did not connect", log.ErrorField(err))
		return
	}
	defer source.Close()

	log.Info("connect dest server", log.String("addr", config.DefaultCliArgs().Addr))
	dest, err := util.NewClient(
		config.DefaultCliArgs().Addr,
		util.WithCliArgs(config.DefaultCliArgs()))
	if err != nil {
		log.Error("did not connect", log.ErrorField(err))
		return
	}
	defer dest.Close()

	req := eventv1.GetEventRequest{
		EventSelector: util.ResolveEvent(arg),
	}

	ctx := context.Background()
	c := eventv1grpc.NewEventServiceClient(source)
	e, err := c.GetEvent(ctx, &req)
	if err != nil {
		log.Error("could not load event", log.ErrorField(err), log.String("event", arg))
		return
	}

	log.Info("Event loaded.",
		log.String("event", e.Event.Name),
		log.Uint32("id", e.Event.Id))

	dp := replay.NewDataProvider(source, e.Event.Id,
		func() *providerv1.RegisterEventRequest {
			recordingMode := func() providerv1.RecordingMode {
				if cfg.DoNotPersist {
					return providerv1.RecordingMode_RECORDING_MODE_DO_NOT_PERSIST
				} else {
					return providerv1.RecordingMode_RECORDING_MODE_PERSIST
				}
			}
			if cfg.EventKey == "" {
				cfg.EventKey = uuid.New().String()
			}
			e.Event.Key = cfg.EventKey
			return &providerv1.RegisterEventRequest{
				Event:         e.Event,
				Track:         e.Track,
				RecordingMode: recordingMode(),
			}
		})
	opts := make([]replay.ReplayOption, 0)
	if cfg.Speed > 0 {
		opts = append(opts, replay.WithSpeed(cfg.Speed))
	}
	if cfg.FastForward != time.Duration(0) {
		opts = append(opts, replay.WithFastForward(cfg.FastForward))
	}
	if cfg.Token != "" {
		opts = append(opts, replay.WithTokenProvider(func() string {
			return cfg.Token
		}))
	}
	opts = append(opts, replay.WithLogging(log.Default()))

	r := replay.NewReplayTask(dest, dp, opts...)
	if err := r.Replay(e.Event.Id); err != nil {
		log.Error("Error replaying event", log.ErrorField(err))
	}
}
