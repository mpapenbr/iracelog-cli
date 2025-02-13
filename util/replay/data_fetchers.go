package replay

import (
	"context"
	"time"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/mpapenbr/iracelog-cli/log"
)

//nolint:whitespace,lll,dupl // by design
func initDriverDataFetcher(
	service racestatev1grpc.RaceStateServiceClient,
	logger *log.Logger,
	eventId uint32,
	lastTS time.Time,
	limit int,
) myFetcher[racestatev1.PublishDriverDataRequest] {
	df := &commonFetcher[racestatev1.PublishDriverDataRequest]{
		lastTS: lastTS,

		loader: func(startTs time.Time) ([]*racestatev1.PublishDriverDataRequest, time.Time, error) {
			if resp, err := service.GetDriverData(context.Background(),
				&racestatev1.GetDriverDataRequest{
					Event: buildEventSelector(eventId),
					Start: buildStartSelector(startTs),
					Num:   int32(limit),
				}); err == nil {
				logger.Debug("loaded driver data",
					log.Int("count", len(resp.DriverData)),
					log.Time("start", startTs),
					log.Time("last", resp.LastTs.AsTime()),
				)
				return resp.DriverData, resp.LastTs.AsTime().Add(time.Millisecond), nil
			} else {
				logger.Error("failed to load driver data", log.ErrorField(err))
				return nil, lastTS, err
			}
		},
	}
	return df
}

//nolint:whitespace,lll,dupl // by design
func initStateDataFetcher(
	service racestatev1grpc.RaceStateServiceClient,
	logger *log.Logger,
	eventId uint32,
	lastTS time.Time,
	limit int,
) myFetcher[racestatev1.PublishStateRequest] {
	df := &commonFetcher[racestatev1.PublishStateRequest]{
		lastTS: lastTS,
		loader: func(startTs time.Time) ([]*racestatev1.PublishStateRequest, time.Time, error) {
			if resp, err := service.GetStates(context.Background(),
				&racestatev1.GetStatesRequest{
					Event: buildEventSelector(eventId),
					Start: buildStartSelector(startTs),
					Num:   int32(limit),
				}); err == nil {
				logger.Debug("loaded state data",
					log.Int("count", len(resp.States)),
					log.Time("start", startTs),
					log.Time("last", resp.LastTs.AsTime()),
				)
				return resp.States, resp.LastTs.AsTime().Add(time.Millisecond), nil
			} else {
				logger.Error("failed to load state data", log.ErrorField(err))
				return nil, lastTS, err
			}
		},
	}

	return df
}

//nolint:whitespace,lll,dupl // by design
func initSpeedmapDataFetcher(
	service racestatev1grpc.RaceStateServiceClient,
	logger *log.Logger,
	eventId uint32,
	lastTS time.Time,
	limit int,
) myFetcher[racestatev1.PublishSpeedmapRequest] {
	df := &commonFetcher[racestatev1.PublishSpeedmapRequest]{
		lastTS: lastTS,
		loader: func(startTs time.Time) ([]*racestatev1.PublishSpeedmapRequest, time.Time, error) {
			if resp, err := service.GetSpeedmaps(context.Background(),
				&racestatev1.GetSpeedmapsRequest{
					Event: buildEventSelector(eventId),
					Start: buildStartSelector(startTs),
					Num:   int32(limit),
				}); err == nil {
				logger.Debug("loaded speedmap data",
					log.Int("count", len(resp.Speedmaps)),
					log.Time("start", startTs),
					log.Time("last", resp.LastTs.AsTime()),
				)
				return resp.Speedmaps, resp.LastTs.AsTime().Add(time.Millisecond), nil
			} else {
				logger.Error("failed to load speedmap data", log.ErrorField(err))
				return nil, lastTS, err
			}
		},
	}

	return df
}

func buildEventSelector(eventId uint32) *commonv1.EventSelector {
	return &commonv1.EventSelector{Arg: &commonv1.EventSelector_Id{Id: int32(eventId)}}
}

func buildStartSelector(t time.Time) *commonv1.StartSelector {
	return &commonv1.StartSelector{
		Arg: &commonv1.StartSelector_RecordStamp{RecordStamp: timestamppb.New(t)},
	}
}

type myFetcher[E any] interface {
	next() *E
}

type (
	myLoaderFunc[E any] func(startTs time.Time) ([]*E, time.Time, error)
	mapToSessionType    func(sessionNum uint32) commonv1.SessionType
)

//nolint:unused // false positive
type commonFetcher[E any] struct {
	loader             myLoaderFunc[E]
	buffer             []*E
	lastTS             time.Time
	resolveSessionType mapToSessionType
}

//nolint:unused // false positive
func (f *commonFetcher[E]) next() *E {
	if len(f.buffer) == 0 {
		f.fetch()
	}
	if len(f.buffer) == 0 {
		return nil
	}
	ret := f.buffer[0]
	f.buffer = f.buffer[1:]

	return ret
}

//nolint:unused // false positive
func (f *commonFetcher[E]) fetch() {
	f.buffer, f.lastTS, _ = f.loader(f.lastTS)
}
