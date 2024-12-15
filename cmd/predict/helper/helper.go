package helper

import (
	"context"
	"fmt"
	"io"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/predict/v1/predictv1grpc"
	predictv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/predict/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/mpapenbr/iracelog-cli/log"
)

type (
	Option        func(*PredictRace)
	ParamProvider func() (*predictv1.PredictParam, error)
	PredictRace   struct {
		client         *grpc.ClientConn
		param          *predictv1.PredictParam
		paramProvider  ParamProvider
		writer         io.Writer
		format         string
		predictService predictv1grpc.PredictServiceClient
		result         *predictv1.PredictResult
	}
)

func NewPredictRace(client *grpc.ClientConn, opts ...Option) (*PredictRace, error) {
	pr := &PredictRace{
		client:         client,
		predictService: predictv1grpc.NewPredictServiceClient(client),
		param:          &predictv1.PredictParam{},
		writer:         nil,
		format:         "text",
	}
	for _, opt := range opts {
		opt(pr)
	}
	var err error
	if pr.paramProvider != nil {
		if pr.param, err = pr.paramProvider(); err != nil {
			return nil, err
		}
	}
	return pr, nil
}

func WithWriter(w io.Writer) Option {
	return func(pr *PredictRace) {
		pr.writer = w
	}
}

func WithFormat(arg string) Option {
	return func(pr *PredictRace) {
		pr.format = arg
	}
}

func WithParamProvider(pp ParamProvider) Option {
	return func(pr *PredictRace) {
		pr.paramProvider = pp
	}
}

func (pr *PredictRace) Predict() error {
	reqPredict := predictv1.PredictRaceRequest{
		Param: pr.param,
	}
	var predictResp *predictv1.PredictRaceResponse
	var err error
	ctx := context.Background()
	if predictResp, err = pr.predictService.PredictRace(ctx, &reqPredict); err != nil {
		log.Error("could not get predict result",
			log.ErrorField(err),
		)
		return err
	}
	pr.result = predictResp.Result
	return nil
}

func (pr *PredictRace) Result() (*predictv1.PredictParam, *predictv1.PredictResult) {
	return pr.param, pr.result
}

func (pr *PredictRace) WriteResult() error {
	if pr.writer != nil {
		var output Output
		switch pr.format {
		case "json":

		case "text":
			output = &textOutput{}
		}
		if output != nil {
			//nolint:errcheck // by design
			pr.writer.Write([]byte(output.Output(pr.param, pr.result)))
		}
	}
	return nil
}

type Output interface {
	Output(param *predictv1.PredictParam, result *predictv1.PredictResult) string
}

type textOutput struct{}

func clearNanos(t *durationpb.Duration) *durationpb.Duration {
	return &durationpb.Duration{Seconds: t.GetSeconds(), Nanos: 0}
}

//nolint:whitespace,lll // by design
func (r textOutput) Output(
	param *predictv1.PredictParam,
	result *predictv1.PredictResult,
) string {
	common := func(p *predictv1.Part) string {
		return fmt.Sprintf("Start: %-9s End: %-9s Dur: %-8s",
			clearNanos(p.Start).AsDuration(),
			clearNanos(p.End).AsDuration(),
			clearNanos(p.Duration).AsDuration())
	}
	ret := fmt.Sprintf("Input:\nRemain Duration: %s SessionTime: %s AvgLap: %s StintLaps:%d\n\n",
		clearNanos(param.Race.Duration).AsDuration(),
		clearNanos(param.Race.Session).AsDuration(),
		param.Stint.AvgLaptime.AsDuration(),
		param.Stint.Lps,
	)
	for _, p := range result.Parts {
		switch p.PartType.(type) {
		case *predictv1.Part_Pit:
			ret += fmt.Sprintf("Pit  : %s\n", common(p))
		case *predictv1.Part_Stint:
			stint := p.GetStint()
			ret += fmt.Sprintf("Stint: %s %3d-%3d (%d)\n",
				common(p), stint.LapStart, stint.LapEnd, stint.Laps)
		default:
			fmt.Println("Unknown part type")
		}
	}
	return ret
}
