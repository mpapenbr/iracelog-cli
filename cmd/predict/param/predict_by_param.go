package param

import (
	"context"
	"os"
	"time"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/predict/v1/predictv1grpc"
	predictv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/predict/v1"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/mpapenbr/iracelog-cli/cmd/predict/cmdopts"
	"github.com/mpapenbr/iracelog-cli/cmd/predict/helper"
	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util"
)

func NewPredictByParamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "param",
		Short: "predict race by provided parameters.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},

		Run: func(cmd *cobra.Command, args []string) {
			predictByParam(cmd.Context())
		},
	}

	cmd.Flags().DurationVar(&cmdopts.RaceDuration, "race-duration", 0,
		"calculate for this race duration (for example: 60m)")
	cmd.Flags().DurationVar(&cmdopts.SessionTime, "session-time", 0,
		"session time as duration where data should begin (for example: 10m)")
	cmd.Flags().DurationVar(&cmdopts.StintAvg, "stint-avg", 0,
		"calc with this average lap time")
	cmd.Flags().DurationVar(&cmdopts.PitOverall, "pit-overall", 0,
		"time used for pitstop")
	cmd.Flags().Int32Var(&cmdopts.StintLaps, "stint-laps", 0,
		"calc with this laps per stint")

	return cmd
}

func predictByParam(ctx context.Context) {
	l := log.GetFromContext(ctx)
	l.Info("connect ism ", log.String("addr", config.DefaultCliArgs().Addr))
	var err error
	conn, err := util.ConnectGrpc(config.DefaultCliArgs())
	if err != nil {
		log.Fatal("did not connect", log.ErrorField(err))
	}
	defer conn.Close()

	pr := newPredictRace(conn)
	if err = pr.predictRace(); err != nil {
		log.Error("could not predict the race", log.ErrorField(err))
	}
}

type predictRace struct {
	client         *grpc.ClientConn
	predictService predictv1grpc.PredictServiceClient
}

func newPredictRace(client *grpc.ClientConn) *predictRace {
	return &predictRace{
		client:         client,
		predictService: predictv1grpc.NewPredictServiceClient(client),
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
	pp := &predictv1.PredictParam{
		Race: &predictv1.RaceParam{
			Duration: durationpb.New(60 * time.Minute),
			Lc:       0,
			Session:  durationpb.New(0),
		},
		Stint: &predictv1.StintParam{
			Lps:        10,
			AvgLaptime: durationpb.New(60 * time.Second),
		},

		Car: &predictv1.CarParam{
			CurrentTrackPos: 0,
			InPit:           false,
			StintLap:        0,
			RemainLapTime:   durationpb.New(0),
		},
		Pit: &predictv1.PitParam{
			Overall: durationpb.New(60 * time.Second),
		},
	}

	cmdopts.MergeOptions(pp)
	return pp, nil
}
