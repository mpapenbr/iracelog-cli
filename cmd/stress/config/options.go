package config

import (
	"time"

	myStress "github.com/mpapenbr/iracelog-cli/stress"
)

var (
	WorkerThreads int
	TestDuration  string
	Pause         string
	Token         string

	PauseDuration     time.Duration
	WorkerProgressArg string
	RampUpDurationArg string
	RampUpIncrease    int
	RampUpInitial     int

	JobLogLevelArg string
)

func CollectStandardJobProcessorOptions() []myStress.OptionFunc {
	ret := []myStress.OptionFunc{}
	ret = append(ret, myStress.WithNumWorker(WorkerThreads))
	if pauseDuration, err := time.ParseDuration(Pause); err == nil {
		ret = append(ret, myStress.WithPauseBetweenRuns(pauseDuration))
	}
	if maxDur, err := time.ParseDuration(TestDuration); err == nil {
		ret = append(ret, myStress.WithMaxDuration(maxDur))
	}
	if workerProgress, err := time.ParseDuration(WorkerProgressArg); err == nil {
		ret = append(ret, myStress.WithWorkerProgress(workerProgress))
	}
	if rampUpDuration, err := time.ParseDuration(RampUpDurationArg); err == nil {
		ret = append(ret, myStress.WithRampUpDuration(rampUpDuration))
	}
	if RampUpInitial > 0 {
		ret = append(ret, myStress.WithRampUpInitialWorkers(RampUpInitial))
	}
	if RampUpIncrease > 0 {
		ret = append(ret, myStress.WithRampUpIncrease(RampUpIncrease))
	}

	return ret
}
