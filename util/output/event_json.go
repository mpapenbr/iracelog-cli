package output

import (
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	eventJson struct{}
)

func (e *eventJson) eventInfo(config *EventOutputConfig) {
	config.outputFunc(protojson.Format(config.eventData.Event))
}

func (e *eventJson) eventCars(config *EventOutputConfig) {
	for _, ce := range config.sortedEntries() {
		config.outputFunc(protojson.Format(ce))
	}
}

func (e *eventJson) eventCarLaps(config *EventOutputConfig) {
	for _, v := range config.eventData.Analysis.CarLaps {
		if config.showCar(v.CarNum) {
			s := protojson.Format(v)
			config.outputFunc(s)
		}
	}
}
