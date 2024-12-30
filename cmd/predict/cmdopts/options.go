package cmdopts

import (
	"time"

	predictv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/predict/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	Event        string
	CarNum       string
	RaceDuration time.Duration
	SessionTime  time.Duration
	PitOverall   time.Duration
	StintAvg     time.Duration
	StintLaps    int32
	// LaptimeSelector predictv1.LaptimeSelector
	LaptimeSelector string
)

func MergeOptions(param *predictv1.PredictParam) {
	if RaceDuration > 0 {
		param.Race.Duration = durationpb.New(RaceDuration)
	}
	if SessionTime > 0 {
		param.Race.Session = durationpb.New(SessionTime)
	}
	if StintAvg > 0 {
		param.Stint.AvgLaptime = durationpb.New(StintAvg)
	}
	if PitOverall > 0 {
		param.Pit.Overall = durationpb.New(PitOverall)
	}
	if StintLaps > 0 {
		param.Stint.Lps = StintLaps
	}
}
