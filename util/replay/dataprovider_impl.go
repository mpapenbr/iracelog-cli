package replay

import (
	"time"

	"buf.build/gen/go/mpapenbr/iracelog/grpc/go/iracelog/racestate/v1/racestatev1grpc"
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	providerv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/provider/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
	"google.golang.org/grpc"

	"github.com/mpapenbr/iracelog-cli/log"
)

type ProvideEventRequest func() *providerv1.RegisterEventRequest

//nolint:whitespace,lll // by design
func NewDataProvider(
	source *grpc.ClientConn,
	eventId uint32,
	eventRequestProvider ProvideEventRequest,
) ReplayDataProvider {
	service := racestatev1grpc.NewRaceStateServiceClient(source)
	getLogger := func(name string) *log.Logger {
		return log.Default().Named("replay").Named(name)
	}
	ret := &dataProviderImpl{
		source:               source,
		eventRequestProvider: eventRequestProvider,
		stateFetcher:         initStateDataFetcher(service, getLogger("state"), eventId, time.Time{}, 100),
		speedmapFetcher:      initSpeedmapDataFetcher(service, getLogger("speedmap"), eventId, time.Time{}, 100),
		driverDataFetcher:    initDriverDataFetcher(service, getLogger("driver"), eventId, time.Time{}, 100),
	}
	return ret
}

type dataProviderImpl struct {
	ReplayDataProvider
	eventSelector        *commonv1.EventSelector
	source               *grpc.ClientConn
	eventRequestProvider ProvideEventRequest
	stateFetcher         myFetcher[racestatev1.PublishStateRequest]
	speedmapFetcher      myFetcher[racestatev1.PublishSpeedmapRequest]
	driverDataFetcher    myFetcher[racestatev1.PublishDriverDataRequest]
}

//nolint:whitespace // false positive
func (d *dataProviderImpl) ProvideEventData(
	eventId uint32,
) *providerv1.RegisterEventRequest {
	if d.eventRequestProvider != nil {
		ret := d.eventRequestProvider()
		d.eventSelector = &commonv1.EventSelector{Arg: &commonv1.EventSelector_Key{
			Key: ret.Event.Key,
		}}
		return ret
	}
	return nil
}

func (r *dataProviderImpl) NextDriverData() *racestatev1.PublishDriverDataRequest {
	item := r.driverDataFetcher.next()
	if item == nil {
		return nil
	}
	item.Event = r.eventSelector
	return item
}

func (r *dataProviderImpl) NextStateData() *racestatev1.PublishStateRequest {
	item := r.stateFetcher.next()
	if item == nil {
		return nil
	}
	item.Event = r.eventSelector
	return item
}

func (r *dataProviderImpl) NextSpeedmapData() *racestatev1.PublishSpeedmapRequest {
	item := r.speedmapFetcher.next()
	if item == nil {
		return nil
	}
	item.Event = r.eventSelector
	return item
}
