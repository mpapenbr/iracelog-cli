package simulate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/livedata/v1/livedatav1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	livedatav1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/livedata/v1"
	"github.com/dustin/go-humanize"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/mpapenbr/iracelog-cli/log"
)

type (
	Option    func(*Webclient)
	Webclient struct {
		ctx                   context.Context
		client                *grpc.ClientConn
		logger                *log.Logger
		live                  livedatav1grpc.LiveDataServiceClient
		wg                    sync.WaitGroup
		stats                 Stats
		statsCallback         func(*Stats)
		statsCallbackDuration time.Duration
		maxErrors             int
	}
	DataStat struct {
		Count uint
		Bytes uint
	}
	Stats struct {
		Analysis DataStat
		Driver   DataStat
		Speedmap DataStat
		State    DataStat
	}
)

func WithContext(ctx context.Context) Option {
	return func(w *Webclient) {
		w.ctx = ctx
	}
}

func WithMaxErrors(maxErrors int) Option {
	return func(w *Webclient) {
		w.maxErrors = maxErrors
	}
}

func WithClient(client *grpc.ClientConn) Option {
	return func(w *Webclient) {
		w.client = client
	}
}

func WithStatsCallback(d time.Duration, callback func(*Stats)) Option {
	return func(w *Webclient) {
		w.statsCallback = callback
		w.statsCallbackDuration = d
	}
}

func NewWebclient(opts ...Option) *Webclient {
	w := &Webclient{
		logger:    log.Default().Named("webclient"),
		wg:        sync.WaitGroup{},
		maxErrors: 5,
	}
	for _, opt := range opts {
		opt(w)
	}
	w.live = livedatav1grpc.NewLiveDataServiceClient(w.client)
	// setup progress report if requested
	if w.statsCallbackDuration > 0 {
		ticker := time.NewTicker(w.statsCallbackDuration)
		go func() {
			for {
				select {
				case <-w.ctx.Done():
					w.logger.Debug("context done")
					ticker.Stop()
					return
				case <-ticker.C:
					w.statsCallback(&w.stats)
				}
			}
		}()
	}
	return w
}

func (w *Webclient) Start(event *commonv1.EventSelector) error {
	w.wg.Add(4)
	go w.liveAnalysis(event)
	go w.liveRaceStates(event)
	go w.liveSpeedmaps(event)
	go w.liveDriverData(event)
	w.logger.Info("waiting for coroutines to finish")
	w.wg.Wait()
	w.logger.Info("coroutines finished")
	return nil
}

func (w *Webclient) GetStats() Stats {
	return w.stats
}

//nolint:dupl,funlen,gocognit,cyclop // by design
func (w *Webclient) liveAnalysis(event *commonv1.EventSelector) {
	defer w.wg.Done()

	myLogger := w.logger.Named("analysis")
	req := livedatav1.LiveAnalysisSelRequest{
		Event: event,
		Selector: &livedatav1.AnalysisSelector{
			Components: []livedatav1.AnalysisComponent{
				livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_COMPUTE_STATES,
				livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_LAPS,
				livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_OCCUPANCIES,
				livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_PITS,
				livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_CAR_STINTS,
				livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_RACE_GRAPH,
				livedatav1.AnalysisComponent_ANALYSIS_COMPONENT_RACE_ORDER,
			},
			CarLapsNumTail:   1,
			RaceGraphNumTail: 1,
		},
	}
	r, err := w.live.LiveAnalysisSel(w.ctx, &req)
	if err != nil {
		myLogger.Error("could not get live data", log.ErrorField(err))
		return
	}
	errorCount := 0
	for {
		select {
		case <-w.ctx.Done():
			myLogger.Debug("context done")
			return
		default:
			resp, err := r.Recv()
			if errors.Is(err, io.EOF) {
				myLogger.Debug("server closed stream")
				return
			}
			st, ok := status.FromError(err)
			if ok {
				//nolint:exhaustive // false positive
				switch st.Code() {
				case codes.DeadlineExceeded, codes.Canceled, codes.Aborted:
					myLogger.Debug("context deadline exceeded")
					return
				case codes.NotFound:
					myLogger.Debug("event may be no longer available")
					return
				}
			}
			if err != nil {
				myLogger.Error("error fetching live analysis", log.ErrorField(err))
				errorCount++
				time.Sleep(1 * time.Second) // just to slow down the error log
				if errorCount > w.maxErrors {
					myLogger.Error("too many errors", log.Int("max", w.maxErrors))
					return
				}
			} else {
				errorCount = 0
				myLogger.Debug("msg rcvd", log.Int("size", proto.Size(resp)))
				w.stats.Analysis.Count++
				w.stats.Analysis.Bytes += uint(proto.Size(resp))
			}
		}
	}
}

//nolint:dupl,gocognit,funlen,cyclop // by design
func (w *Webclient) liveRaceStates(event *commonv1.EventSelector) {
	defer w.wg.Done()

	myLogger := w.logger.Named("states")
	req := livedatav1.LiveRaceStateRequest{Event: event}

	r, err := w.live.LiveRaceState(w.ctx, &req)
	if err != nil {
		myLogger.Error("could not get live data", log.ErrorField(err))
		return
	}
	errorCount := 0
	for {
		select {
		case <-w.ctx.Done():
			myLogger.Debug("context done")
			return
		default:
			resp, err := r.Recv()
			if errors.Is(err, io.EOF) {
				myLogger.Debug("server closed stream")
				return
			}
			st, ok := status.FromError(err)
			if ok {
				//nolint:exhaustive // false positive
				switch st.Code() {
				case codes.DeadlineExceeded, codes.Canceled, codes.Aborted:
					myLogger.Debug("context deadline exceeded")
					return
				case codes.NotFound:
					myLogger.Debug("event may be no longer available")
					return
				}
			}
			if err != nil {
				myLogger.Error("error fetching live state", log.ErrorField(err))
				errorCount++
				time.Sleep(1 * time.Second) // just to slow down the error log
				if errorCount > w.maxErrors {
					myLogger.Error("too many errors", log.Int("max", w.maxErrors))
					return
				}
			} else {
				errorCount = 0
				myLogger.Debug("msg rcvd", log.Int("size", proto.Size(resp)))
				w.stats.State.Count++
				w.stats.State.Bytes += uint(proto.Size(resp))
			}
		}
	}
}

//nolint:dupl,gocognit,funlen,cyclop // by design
func (w *Webclient) liveSpeedmaps(event *commonv1.EventSelector) {
	defer w.wg.Done()

	myLogger := w.logger.Named("speedmaps")
	req := livedatav1.LiveSpeedmapRequest{Event: event}

	r, err := w.live.LiveSpeedmap(w.ctx, &req)
	if err != nil {
		myLogger.Error("could not get live data", log.ErrorField(err))
		return
	}
	errorCount := 0
	for {
		select {
		case <-w.ctx.Done():
			myLogger.Debug("context done")
			return
		default:
			resp, err := r.Recv()
			if errors.Is(err, io.EOF) {
				myLogger.Debug("server closed stream")
				return
			}
			st, ok := status.FromError(err)
			if ok {
				//nolint:exhaustive // false positive
				switch st.Code() {
				case codes.DeadlineExceeded, codes.Canceled, codes.Aborted:
					myLogger.Debug("context deadline exceeded")
					return
				case codes.NotFound:
					myLogger.Debug("event may be no longer available")
					return
				}
			}
			if err != nil {
				myLogger.Error("error fetching live speedmaps", log.ErrorField(err))
				errorCount++
				time.Sleep(1 * time.Second) // just to slow down the error log
				if errorCount > w.maxErrors {
					myLogger.Error("too many errors", log.Int("max", w.maxErrors))
					return
				}
			} else {
				errorCount = 0
				myLogger.Debug("msg rcvd", log.Int("size", proto.Size(resp)))
				w.stats.Speedmap.Count++
				w.stats.Speedmap.Bytes += uint(proto.Size(resp))
			}
		}
	}
}

//nolint:gocognit,funlen,cyclop // by design
func (w *Webclient) liveDriverData(event *commonv1.EventSelector) {
	defer w.wg.Done()

	myLogger := w.logger.Named("driver")
	req := livedatav1.LiveDriverDataRequest{Event: event}

	r, err := w.live.LiveDriverData(w.ctx, &req)
	if err != nil {
		myLogger.Error("could not get live data", log.ErrorField(err))
		return
	}
	errorCount := 0
	for {
		select {
		case <-w.ctx.Done():
			myLogger.Debug("context done")
			return
		default:
			resp, err := r.Recv()
			if errors.Is(err, io.EOF) {
				myLogger.Debug("server closed stream")
				return
			}
			st, ok := status.FromError(err)
			if ok {
				//nolint:exhaustive // false positive
				switch st.Code() {
				case codes.DeadlineExceeded, codes.Canceled, codes.Aborted:
					myLogger.Debug("context deadline exceeded", log.Int("code", int(st.Code())))
					return

				case codes.OK:
					myLogger.Debug("msg rcvd", log.Int("size", proto.Size(resp)))
					w.stats.Driver.Count++
					w.stats.Driver.Bytes += uint(proto.Size(resp))
					errorCount = 0
				default:
					myLogger.Error("error fetching live driver data", log.ErrorField(err))
					errorCount++
					if errorCount > w.maxErrors {
						myLogger.Error("too many errors",
							log.Int("errorCount", errorCount),
							log.Int("max", w.maxErrors))
						return
					}
					time.Sleep(1 * time.Second) // just to slow down the error log
				}
			}
		}
	}
}

func (ds *DataStat) Add(other *DataStat) {
	ds.Count += other.Count
	ds.Bytes += other.Bytes
}

func (ds *DataStat) String() string {
	return fmt.Sprintf("%d/%s", ds.Count, humanize.IBytes(uint64(ds.Bytes)))
}

func (s *Stats) Add(other *Stats) {
	s.Analysis.Add(&other.Analysis)
	s.Driver.Add(&other.Driver)
	s.Speedmap.Add(&other.Speedmap)
	s.State.Add(&other.State)
}

func (s *Stats) String() string {
	return fmt.Sprintf("Analysis: %s, Driver: %s, Speedmap: %s, State: %s",
		s.Analysis.String(), s.Driver.String(), s.Speedmap.String(), s.State.String())
}
