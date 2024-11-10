package config

import (
	"time"

	myStress "github.com/mpapenbr/iracelog-cli/stress"
)

var (
	WorkerThreads int
	TestDuration  time.Duration
	Pause         time.Duration
	Token         string

	PauseDuration  time.Duration
	WorkerProgress time.Duration
	RampUpDuration time.Duration
	RampUpIncrease int
	RampUpInitial  int

	JobLogLevelArg string
)

func CollectStandardJobProcessorOptions() []myStress.OptionFunc {
	ret := []myStress.OptionFunc{
		myStress.WithNumWorker(WorkerThreads),
		myStress.WithPauseBetweenRuns(Pause),
		myStress.WithMaxDuration(TestDuration),
		myStress.WithWorkerProgress(WorkerProgress),
		myStress.WithRampUpDuration(RampUpDuration),
	}

	if RampUpInitial > 0 {
		ret = append(ret, myStress.WithRampUpInitialWorkers(RampUpInitial))
	}
	if RampUpIncrease > 0 {
		ret = append(ret, myStress.WithRampUpIncrease(RampUpIncrease))
	}

	return ret
}
