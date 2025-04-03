package event

import (
	"fmt"
	"slices"

	carv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/car/v1"
	eventv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/event/v1"

	"github.com/mpapenbr/iracelog-cli/util/output"
)

type (
	Option            func(*EventOutputConfig)
	EventOutputConfig struct {
		eventData    *eventv1.GetEventResponse
		carNumFilter []string
		format       output.Format
		components   []Component
		outputFunc   func(s string)
	}
	EventOutput interface {
		Output()
	}

	eventOutput struct {
		config    *EventOutputConfig
		outputter formatOutput
	}
	formatOutput interface {
		eventInfo(config *EventOutputConfig)
		eventCars(config *EventOutputConfig)
		eventCarLaps(config *EventOutputConfig)
	}
)

func NewEventOutput(eventData *eventv1.GetEventResponse, opts ...Option) EventOutput {
	ret := &EventOutputConfig{
		eventData:  eventData,
		outputFunc: func(s string) { fmt.Println(s) },
		format:     output.FormatText,
		components: []Component{},
	}
	for _, opt := range opts {
		opt(ret)
	}
	//nolint:exhaustive // by design
	switch ret.format {
	case output.FormatText:
		return &eventOutput{config: ret, outputter: &eventText{}}
	case output.FormatJSON:
		return &eventOutput{config: ret, outputter: &eventJSON{}}
	}
	return &eventOutput{config: ret, outputter: &eventEmpty{}}
}

func (e *eventOutput) Output() {
	if len(e.config.components) == 0 {
		e.config.components = []Component{
			ComponentEventInfo,
			ComponentEventCars,
			ComponentEventCarLaps,
		}
	}
	for _, c := range e.config.components {
		switch c {
		case ComponentEventInfo:
			e.outputter.eventInfo(e.config)
		case ComponentEventCars:
			e.outputter.eventCars(e.config)
		case ComponentEventCarLaps:
			e.outputter.eventCarLaps(e.config)
		case ComponentUnknown:
		}
	}
}

func WithCarNumFilter(carNumFilter []string) Option {
	return func(e *EventOutputConfig) {
		e.carNumFilter = carNumFilter
	}
}

func WithComponents(comps []Component) Option {
	return func(e *EventOutputConfig) {
		e.components = comps
	}
}

func WithFormat(f output.Format) Option {
	return func(e *EventOutputConfig) {
		e.format = f
	}
}

func (e *EventOutputConfig) sortedEntries() []*carv1.CarEntry {
	return slices.SortedFunc(slices.Values(e.eventData.Car.Entries),
		func(i, j *carv1.CarEntry) int {
			return int(i.Car.CarNumberRaw) - int(j.Car.CarNumberRaw)
		})
}

func (e *EventOutputConfig) showCar(carNum string) bool {
	if len(e.carNumFilter) == 0 {
		return true
	}
	return slices.Contains(e.carNumFilter, carNum)
}
