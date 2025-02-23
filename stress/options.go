package stress

import (
	"time"
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

func CollectStandardJobProcessorOptions() []OptionFunc {
	ret := []OptionFunc{
		WithNumWorker(WorkerThreads),
		WithPauseBetweenRuns(Pause),
		WithMaxDuration(TestDuration),
		WithWorkerProgress(WorkerProgress),
		WithRampUpDuration(RampUpDuration),
	}

	if RampUpInitial > 0 {
		ret = append(ret, WithRampUpInitialWorkers(RampUpInitial))
	}
	if RampUpIncrease > 0 {
		ret = append(ret, WithRampUpIncrease(RampUpIncrease))
	}

	return ret
}
