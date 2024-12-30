package event

import (
	"context"
	"errors"
	"os"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/predict/v1/predictv1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	predictv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/predict/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/mpapenbr/iracelog-cli/cmd/predict/cmdopts"
	"github.com/mpapenbr/iracelog-cli/cmd/predict/helper"
	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewPredictEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "predict stored event.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			predictEvent(cmd.Context(), args[0])
		},
	}

	cmd.Flags().DurationVar(&cmdopts.RaceDuration, "race-duration", 0,
		"calculate for this race duration (for example: 60m)")
	cmd.Flags().DurationVar(&cmdopts.SessionTime, "session-time", 0,
		"session time as duration where data should begin (for example: 10m)")
	cmd.Flags().StringVar(&cmdopts.LaptimeSelector, "laptime-selector", "prev-stint-avg",
		"which laptime should be used for prediction")

	cmd.Flags().DurationVar(&cmdopts.StintAvg, "stint-avg", 0,
		"calc with this average lap time")
	cmd.Flags().DurationVar(&cmdopts.PitOverall, "pit-overall", 0,
		"time used for pitstop")
	cmd.Flags().Int32Var(&cmdopts.StintLaps, "stint-laps", 0,
		"calc with this laps per stint")
	cmd.Flags().StringVar(&cmdopts.CarNum,
		"carnum", "", "predict for this car number")

	//nolint:errcheck // by design
	cmd.MarkFlagRequired("carnum")
	//nolint:errcheck // by design
	cmd.MarkFlagRequired("session-time")
	return cmd
}

func predictEvent(ctx context.Context, event string) {
	l := log.GetFromContext(ctx)
	l.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	var err error
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()

	pr := newPredictRace(conn, event)
	if err = pr.predictRace(); err != nil {
		log.Error("could not predict the race", log.ErrorField(err))
	}
}

type predictRace struct {
	client         *grpc.ClientConn
	predictService predictv1grpc.PredictServiceClient
	event          string
}

func newPredictRace(client *grpc.ClientConn, event string) *predictRace {
	return &predictRace{
		client:         client,
		predictService: predictv1grpc.NewPredictServiceClient(client),
		event:          event,
	}
}

//nolint:gocritic // by design
func (pr *predictRace) predictRace() error {
	var err error
	var p *helper.PredictRace
	if p, err = helper.NewPredictRace(pr.client,
		helper.WithWriter(os.Stdout),
		helper.WithParamProvider(func() (*predictv1.PredictParam, error) {
			return pr.provideParam()
		})); err != nil {
		return err
	}
	if err = p.Predict(); err != nil {
		return err
	}
	if err = p.WriteResult(); err != nil {
		return err
	}
	return nil
}

func (pr *predictRace) provideParam() (*predictv1.PredictParam, error) {
	var eventSel *commonv1.EventSelector
	var startSel *commonv1.StartSelector
	var laptimeSel predictv1.LaptimeSelector
	var err error
	eventSel = util.ResolveEvent(pr.event)
	if eventSel == nil {
		log.Error("could not resolve event", log.String("event", pr.event))
		return nil, errors.New("could not resolve event")
	}
	if startSel, err = util.ResolveStartSelector(cmdopts.SessionTime, ""); err != nil {
		log.Error("could not resolve session time",
			log.Duration("session-time", cmdopts.SessionTime))
		return nil, err
	}
	log.Debug("resolved start selector", log.Any("start-selector", startSel))
	//nolint:lll // readability
	if laptimeSel, err = cmdopts.ParseLaptimeSelector(cmdopts.LaptimeSelector); err != nil {
		log.Error("could not resolve laptime selector",
			log.String("laptime-selector", cmdopts.LaptimeSelector))
		return nil, err
	}

	req := predictv1.GetPredictParamRequest{
		EventSelector:   eventSel,
		StartSelector:   startSel,
		CarNum:          cmdopts.CarNum,
		LaptimeSelector: laptimeSel,
	}
	var resp *predictv1.GetPredictParamResponse
	if resp, err = pr.predictService.GetPredictParam(
		context.Background(), &req); err != nil {
		log.Error("could not get predict parameter",
			log.ErrorField(err),
			log.String("event", pr.event))
		return nil, err
	}
	log.Debug("Predict parameter retrieved.")
	cmdopts.MergeOptions(resp.Param)
	return resp.Param, nil
}
