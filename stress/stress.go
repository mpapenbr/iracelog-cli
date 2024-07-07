package stress

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/mpapenbr/iracelog-cli/log"
)

type Job struct {
	Id           int              // used for overall id
	WorkerId     int              // used to identify worker this job is assigned to
	TargetClient *grpc.ClientConn // used for communication with target backend
	SourceClient *grpc.ClientConn // used for communication with source backend (if needed)
	Ctx          context.Context  // used for cancellation
}

type JobResult struct {
	TimeUsed time.Duration // time
	Error    error
	Request  *Job // reference to request
}

type JobError struct {
	JobId int
	Error error
}

type WorkerStats struct {
	Id       int
	JobsDone int
	Errors   []JobError
	TimeUsed time.Duration
}

type (
	JobHandler func(j *Job) error
)

type JobProcessor struct {
	numWorker            int
	pause                time.Duration
	duration             time.Duration // max time the JobProcessor is running
	wgWorker             sync.WaitGroup
	wgResult             sync.WaitGroup
	queue                chan *Job
	results              chan *JobResult
	doSchedule           bool
	pLogger              *log.Logger // processor logger
	wLogger              *log.Logger // worker logger
	jobHandler           JobHandler
	workerProgress       time.Duration // show worker progress if > 0
	targetClientProvider func() *grpc.ClientConn
	sourceClientProvider func() *grpc.ClientConn

	// collector   dvlResultsCollector
	workerStats []WorkerStats
	nextJobId   int
}

type OptionFunc func(sp *JobProcessor)

func WithJobHandler(handler JobHandler) OptionFunc {
	return func(sp *JobProcessor) {
		sp.jobHandler = handler
	}
}

func WithNumWorker(numWorker int) OptionFunc {
	return func(sp *JobProcessor) {
		sp.numWorker = numWorker
	}
}

func WithPauseBetweenRuns(pause time.Duration) OptionFunc {
	return func(sp *JobProcessor) {
		sp.pause = pause
	}
}

func WithMaxDuration(duration time.Duration) OptionFunc {
	return func(sp *JobProcessor) {
		sp.duration = duration
	}
}

func WithWorkerProgress(duration time.Duration) OptionFunc {
	return func(sp *JobProcessor) {
		sp.workerProgress = duration
	}
}

func WithLogging(logger *log.Logger) OptionFunc {
	return func(sp *JobProcessor) {
		sp.wLogger = log.GetLoggerManager().GetLogger("stress.worker")
		sp.pLogger = log.GetLoggerManager().GetLogger("stress.processor")
	}
}

func WithTargetClientProvider(provider func() *grpc.ClientConn) OptionFunc {
	return func(sp *JobProcessor) {
		sp.targetClientProvider = provider
	}
}

func WithSourceClientProvider(provider func() *grpc.ClientConn) OptionFunc {
	return func(sp *JobProcessor) {
		sp.sourceClientProvider = provider
	}
}

func NewJobProcessor(opts ...OptionFunc) *JobProcessor {
	ret := &JobProcessor{
		numWorker:  1,
		pause:      time.Second,
		duration:   time.Minute * 10,
		wgWorker:   sync.WaitGroup{},
		wgResult:   sync.WaitGroup{},
		queue:      make(chan *Job),
		results:    make(chan *JobResult),
		doSchedule: true,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

//nolint:funlen // by design
func (p *JobProcessor) Run() {
	// this is used to cancel the processor and resultCollector
	ctx, cancel := context.WithCancel(context.Background())
	// a different context is used for the workers
	// to ensure they are terminated at the deadline
	workerCtx, workerCtxCancel := context.WithTimeout(context.Background(), p.duration)
	defer workerCtxCancel()
	deadline, _ := workerCtx.Deadline()
	p.pLogger.Info("Processor deadline", log.Time("deadline", deadline))

	// setup result collector
	go p.resultCollector(ctx)

	p.pLogger.Info("initialize worker", log.Int("worker", p.numWorker))
	for i := 0; i < p.numWorker; i++ {
		p.wgWorker.Add(1)
		workerStats := WorkerStats{Id: i}
		p.workerStats = append(p.workerStats, workerStats)
		go p.jobWorker(workerStats, workerCtx)
	}

	// create initial jobs and add them to the queue
	for p.nextJobId = 1; p.nextJobId <= p.numWorker; p.nextJobId++ {
		p.queue <- &Job{Id: p.nextJobId}
	}

	// setup worker progress report if requested
	if p.workerProgress > 0 {
		ticker := time.NewTicker(p.workerProgress)
		go p.logWorkerProgress(ticker, ctx)
	}

	// setup timer to stop the stress test

	go func() {
		p.pLogger.Info("processing time", log.Duration("duration", p.duration))
		time.Sleep(p.duration)
		p.pLogger.Debug("Signaling reschedule stop")
		p.doSchedule = false

		p.pLogger.Debug("Waiting for outstanding results")
		p.wgResult.Wait()

		p.pLogger.Debug("Signaling cancel")
		cancel()
	}()

	p.pLogger.Debug("Waiting for jobs to terminate")
	p.wgWorker.Wait()
	p.pLogger.Info("All jobs finished")
}

//nolint:gocognit // false positive
func (p *JobProcessor) resultCollector(ctx context.Context) {
	collected := 0
	for {
		select {
		case <-ctx.Done():
			p.pLogger.Info("maxDuration reached, terminating collector")
			return

		case result := <-p.results:
			collected++
			p.pLogger.Debug("Got result from job",
				log.Int("jobId", result.Request.Id),
				log.Int("worker", result.Request.WorkerId),
				log.Int("collected", collected),
			)

			ws := &p.workerStats[result.Request.WorkerId]
			ws.JobsDone++
			if result.Error != nil {
				ws.Errors = append(ws.Errors, JobError{
					JobId: result.Request.Id, Error: result.Error,
				})
			}
			p.wgResult.Done()

			if p.doSchedule {
				go func() {
					if p.pause > 0 {
						//nolint:gosec // false positive
						pauseDur := time.Duration(rand.Intn(int(p.pause)))
						p.pLogger.Debug("pausing before next run", log.Duration("pause", pauseDur))
						time.Sleep(pauseDur)
					}
					if p.doSchedule {
						p.pLogger.Debug("about to issue next job", log.Int("jobId", p.nextJobId))
						p.queue <- &Job{Id: p.nextJobId}
						p.nextJobId++
					} else {
						p.pLogger.Debug("NOT issuing next job, time is up", log.Int("jobId", p.nextJobId))
					}
				}()
			}
		}
	}
}

//nolint:whitespace // false positive
func (p *JobProcessor) jobWorker(
	workerStats WorkerStats,
	ctx context.Context,
) {
	defer p.wgWorker.Done()
	var targetClient *grpc.ClientConn
	var sourceClient *grpc.ClientConn
	if p.targetClientProvider != nil {
		targetClient = p.targetClientProvider()
	}
	if p.sourceClientProvider != nil {
		sourceClient = p.sourceClientProvider()
	}
	for {
		select {
		case <-ctx.Done():
			// used for terminating the job when time is up
			return
		case job := <-p.queue:
			job.WorkerId = workerStats.Id
			job.TargetClient = targetClient
			job.SourceClient = sourceClient
			job.Ctx = ctx
			p.executeJob(job)
		}
	}
}

func (p *JobProcessor) executeJob(j *Job) {
	p.wgResult.Add(1)
	start := time.Now()
	err := p.jobHandler(j)
	p.results <- &JobResult{TimeUsed: time.Since(start), Error: err, Request: j}
}

func (p *JobProcessor) logWorkerProgress(ticker *time.Ticker, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			p.pLogger.Debug("testDuration reached, terminating workerProgress")
			ticker.Stop()
			return
		case <-ticker.C:
			p.pLogger.Debug("About to show progress of workers")

			//nolint:gocritic // false positive
			for _, item := range p.workerStats {
				p.wLogger.Info("progress",
					log.Int("worker", item.Id),
					log.Int("jobsDone", item.JobsDone),
					log.Int("errors", len(item.Errors)),
				)
			}
		}
	}
}
