package event

import (
	"google.golang.org/protobuf/encoding/protojson"
)

type (
	eventJSON struct{}
)

func (e *eventJSON) eventInfo(config *EventOutputConfig) {
	config.outputFunc(protojson.Format(config.eventData.Event))
}

func (e *eventJSON) eventCars(config *EventOutputConfig) {
	for _, ce := range config.sortedEntries() {
		config.outputFunc(protojson.Format(ce))
	}
}

func (e *eventJSON) eventCarLaps(config *EventOutputConfig) {
	for _, v := range config.eventData.Analysis.CarLaps {
		if config.showCar(v.CarNum) {
			s := protojson.Format(v)
			config.outputFunc(s)
		}
	}
}
