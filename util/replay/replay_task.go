package replay

import (
	"context"
	"sync"
	"time"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/provider/v1/providerv1grpc"
	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"
	providerv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/provider/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/mpapenbr/iracelog-cli/log"
)

type ReplayDataProvider interface {
	ProvideEventData(eventId uint32) *providerv1.RegisterEventRequest
	NextDriverData() *racestatev1.PublishDriverDataRequest
	NextStateData() *racestatev1.PublishStateRequest
	NextSpeedmapData() *racestatev1.PublishSpeedmapRequest
}
type ReplayOption func(*ReplayTask)

//nolint:whitespace // false positive
func NewReplayTask(
	dest *grpc.ClientConn,
	dataProvider ReplayDataProvider,
	opts ...ReplayOption,
) *ReplayTask {
	ret := &ReplayTask{
		dataProvider: dataProvider,
		dest:         dest,
		myLog:        log.Default(),
		ctx:          context.Background(),
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func WithFastForward(ff time.Duration) ReplayOption {
	return func(r *ReplayTask) {
		r.fastForward = ff
	}
}

func WithTokenProvider(provider func() string) ReplayOption {
	return func(r *ReplayTask) {
		r.tokenProvider = provider
	}
}

func WithSpeed(speed int) ReplayOption {
	return func(r *ReplayTask) {
		r.speed = speed
	}
}

func WithContext(ctx context.Context) ReplayOption {
	return func(r *ReplayTask) {
		r.ctx = ctx
	}
}

func WithLogging(logger *log.Logger) ReplayOption {
	return func(r *ReplayTask) {
		r.myLog = logger.Named("replay")
	}
}

type ReplayTask struct {
	dataProvider ReplayDataProvider
	dest         *grpc.ClientConn // destination server

	ctx              context.Context
	localCtx         context.Context
	localCancel      context.CancelFunc
	providerService  providerv1grpc.ProviderServiceClient
	raceStateService racestatev1grpc.RaceStateServiceClient
	event            *eventv1.Event

	wg             sync.WaitGroup
	stateChan      chan *racestatev1.PublishStateRequest
	speedmapChan   chan *racestatev1.PublishSpeedmapRequest
	driverDataChan chan *racestatev1.PublishDriverDataRequest
	fastForward    time.Duration
	ffStopTime     time.Time // time when fast forward should stop
	tokenProvider  func() string
	speed          int
	myLog          *log.Logger // used to for replay task related logging
}

func (p *peekDriverData) ts() time.Time {
	return p.dataReq.Timestamp.AsTime()
}

func (p *peekDriverData) publish() error {
	p.logger.Debug("Sending driver data", log.Time("time", p.dataReq.Timestamp.AsTime()))
	ctx := p.r.prepOutgoingContext(p.r.ctx)
	if _, err := p.r.raceStateService.PublishDriverData(ctx, p.dataReq); err != nil {
		return err
	}
	return nil
}

func (p *peekStateData) ts() time.Time {
	return p.dataReq.Timestamp.AsTime()
}

func (p *peekStateData) publish() error {
	p.logger.Debug("Sending state data", log.Time("time", p.dataReq.Timestamp.AsTime()))
	ctx := p.r.prepOutgoingContext(p.r.ctx)
	if _, err := p.r.raceStateService.PublishState(ctx, p.dataReq); err != nil {
		return err
	}
	return nil
}

// --- SpeedmapData ---
func (p *peekSpeedmapData) ts() time.Time {
	if p.dataReq == nil {
		return time.Unix(0, 0).Add(1<<63 - 1) // very far in the future
	}

	return p.dataReq.Timestamp.AsTime()
}

func (p *peekSpeedmapData) publish() error {
	p.logger.Debug("Sending speedmap data", log.Time("time", p.dataReq.Timestamp.AsTime()))
	ctx := p.r.prepOutgoingContext(p.r.ctx)
	if _, err := p.r.raceStateService.PublishSpeedmap(ctx, p.dataReq); err != nil {
		return err
	}
	return nil
}

func (r *ReplayTask) Replay(eventId uint32) error {
	r.providerService = providerv1grpc.NewProviderServiceClient(r.dest)
	r.raceStateService = racestatev1grpc.NewRaceStateServiceClient(r.dest)
	r.localCtx, r.localCancel = context.WithCancel(context.Background())
	defer r.localCancel()

	r.stateChan = make(chan *racestatev1.PublishStateRequest)
	r.speedmapChan = make(chan *racestatev1.PublishSpeedmapRequest)
	r.driverDataChan = make(chan *racestatev1.PublishDriverDataRequest)

	var err error
	registerReq := r.dataProvider.ProvideEventData(eventId)

	if r.event, err = r.registerEvent(registerReq); err != nil {
		return err
	}
	r.myLog.Info("replaying event",
		log.Uint32("id", eventId),
		log.String("key", r.event.Key),
		log.String("event", r.event.Name),
	)

	r.wg = sync.WaitGroup{}
	r.wg.Add(4)
	go r.provideDriverData()
	go r.provideStateData()
	go r.provideSpeedmapData()
	go r.sendData()

	r.myLog.Debug("Waiting for tasks to finish")
	r.wg.Wait()
	r.myLog.Debug("About to unregister event")
	err = r.unregisterEvent()
	r.myLog.Debug("Event unregistered", log.String("key", r.event.Key))

	return err
}

//nolint:funlen,gocognit,cyclop //  by design
func (r *ReplayTask) sendData() {
	defer r.wg.Done()

	pData := make([]peek, 0, 3)
	pData = append(pData, //
		&peekStateData{
			commonStateData[racestatev1.PublishStateRequest]{
				r: r, dataChan: r.stateChan, providerType: StateData,
				logger: r.myLog.Named("state"),
			},
		},
		&peekDriverData{
			commonStateData[racestatev1.PublishDriverDataRequest]{
				r: r, dataChan: r.driverDataChan, providerType: DriverData,
				logger: r.myLog.Named("driver"),
			},
		},
		&peekSpeedmapData{
			commonStateData[racestatev1.PublishSpeedmapRequest]{
				r: r, dataChan: r.speedmapChan, providerType: SpeedmapData,
				logger: r.myLog.Named("speedmap"),
			},
		},
	)
	// init the peek data provider and check if they are exhausted
	init := []peek{}
	for _, p := range pData {
		if !p.refill() {
			r.myLog.Debug("initial exhausted",
				log.String("provider", string(p.provider())))
		} else {
			init = append(init, p)
		}
	}
	pData = init
	lastTs := time.Time{}

	for {
		var selector providerType
		var current peek
		var delta time.Duration
		var currentIdx int
		// create a max time from  (don't use time.Unix(1<<63-1), that's not what we want)
		nextTs := time.Unix(0, 0).Add(1<<63 - 1)

		for i, p := range pData {
			if p.ts().Before(nextTs) {
				nextTs = p.ts()
				selector = p.provider()
				current = pData[i]
				currentIdx = i
			}
		}
		r.computeFastForwardStop(nextTs)

		if !lastTs.IsZero() {
			wait := r.calcWaitTime(nextTs, lastTs)
			if wait > 0 {
				r.myLog.Debug("Sleeping",
					log.Time("time", nextTs),
					log.Duration("delta", delta),
					log.Duration("wait", wait),
				)
				time.Sleep(wait)
			}
		}
		lastTs = nextTs
		if current == nil {
			r.myLog.Error("No provider found")
			return
		}
		if err := current.publish(); err != nil {
			st, ok := status.FromError(err)
			if ok {
				//nolint:exhaustive // false positive
				switch st.Code() {
				case codes.DeadlineExceeded, codes.Canceled, codes.Aborted:
					r.myLog.Debug("context deadline exceeded")
					return
				}
			}
			r.myLog.Error("Error publishing data", log.ErrorField(err))
			return
		}
		if !current.refill() {
			r.myLog.Debug("exhausted", log.String("provider", string(selector)))
			pData = append(pData[:currentIdx], pData[currentIdx+1:]...)
			if len(pData) == 0 {
				r.myLog.Debug("All providers exhausted")
				return
			}
		}
	}
}

func (r *ReplayTask) computeFastForwardStop(cur time.Time) {
	if r.fastForward > 0 && r.ffStopTime.IsZero() {
		r.ffStopTime = cur.Add(r.fastForward)
		r.myLog.Debug("Fast forward stop time set", log.Time("time", r.ffStopTime))
	}
}

func (r *ReplayTask) calcWaitTime(nextTs, lastTs time.Time) time.Duration {
	delta := nextTs.Sub(lastTs)

	// handle fast forward
	if nextTs.Before(r.ffStopTime) {
		return 0
	}
	if r.speed > 0 {
		return time.Duration(int(delta.Nanoseconds()) / r.speed)
	} else {
		return delta
	}
}

func (r *ReplayTask) provideDriverData() {
	defer r.wg.Done()
	i := 0
	for {
		item := r.dataProvider.NextDriverData()
		if item == nil {
			r.myLog.Debug("No more driver data")
			close(r.driverDataChan)
			return
		}
		select {
		case r.driverDataChan <- item:
			i++
			r.myLog.Debug("Sent data on driverDataChen", log.Int("i", i))
		case <-r.ctx.Done():
			close(r.driverDataChan)
			r.myLog.Debug("Context done (inner)")
			return
		}
	}
}

func (r *ReplayTask) provideStateData() {
	defer r.wg.Done()
	i := 0
	for {
		item := r.dataProvider.NextStateData()
		if item == nil {
			r.myLog.Debug("No more state data")
			close(r.stateChan)
			return
		}
		select {
		case r.stateChan <- item:
			i++
			r.myLog.Debug("Sent data on stateDataChan", log.Int("i", i))
		case <-r.ctx.Done():
			close(r.stateChan)
			r.myLog.Debug("Context done (inner)")
			return
		}
	}
}

func (r *ReplayTask) provideSpeedmapData() {
	defer r.wg.Done()
	i := 0
	for {
		item := r.dataProvider.NextSpeedmapData()
		if item == nil {
			r.myLog.Debug("No more speedmap data")
			close(r.speedmapChan)
			return
		}
		select {
		case r.speedmapChan <- item:
			i++
			r.myLog.Debug("Sent data on speedmapChan", log.Int("i", i))
		case <-r.ctx.Done():
			close(r.speedmapChan)
			r.myLog.Debug("Context done (inner)")
			return
		}
	}
}

//nolint:whitespace // by design
func (r *ReplayTask) registerEvent(eventReq *providerv1.RegisterEventRequest) (
	*eventv1.Event, error,
) {
	resp, err := r.providerService.RegisterEvent(r.prepOutgoingContext(r.ctx), eventReq)
	if err == nil {
		return resp.Event, nil
	}
	return nil, err
}

func (r *ReplayTask) unregisterEvent() error {
	req := &providerv1.UnregisterEventRequest{
		EventSelector: r.buildEventSelector(),
	}
	_, err := r.providerService.UnregisterEvent(
		r.prepOutgoingContext(context.Background()),
		req)
	return err
}

func (r *ReplayTask) buildEventSelector() *commonv1.EventSelector {
	return &commonv1.EventSelector{Arg: &commonv1.EventSelector_Key{Key: r.event.Key}}
}

// helper to add the api-token to the outgoing context
func (r *ReplayTask) prepOutgoingContext(ctx context.Context) context.Context {
	if r.tokenProvider != nil {
		md := metadata.Pairs("api-token", r.tokenProvider())
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}
