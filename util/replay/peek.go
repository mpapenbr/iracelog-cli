package replay

import (
	"time"

	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"

	"github.com/mpapenbr/iracelog-cli/log"
)

type providerType string

const (
	DriverData   providerType = "DriverData"
	StateData    providerType = "StateData"
	SpeedmapData providerType = "SpeedmapData"
)

type stampInfo struct {
	sessionType commonv1.SessionType
	ts          time.Time
}

// this is used to peek into the data stream of the different provider.
// we need this to create an order of messages to be published
// therefore we use the timestamp of the provided messages
type peek interface {
	stamp() *stampInfo
	provider() providerType
	publish() error
	refill() bool
}
type commonStateData[E any] struct {
	dataChan     chan *E
	dataReq      *E
	r            *ReplayTask
	providerType providerType
	logger       *log.Logger
	mapFunc      mapToSessionType
}

func (p *commonStateData[E]) refill() bool {
	var ok bool
	p.dataReq, ok = <-p.dataChan
	return ok
}

func (p *commonStateData[E]) provider() providerType {
	return p.providerType
}

type peekStateData struct {
	commonStateData[racestatev1.PublishStateRequest]
}
type peekSpeedmapData struct {
	commonStateData[racestatev1.PublishSpeedmapRequest]
}
type peekDriverData struct {
	commonStateData[racestatev1.PublishDriverDataRequest]
}
