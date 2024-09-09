package output

import (
	"fmt"

	analysisv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/analysis/v1"
)

type (
	eventText struct{}
)

func (e *eventText) eventInfo(config *EventOutputConfig) {
	config.outputFunc(fmt.Sprintf("Event: %s", config.eventData.Event.Name))
	if config.eventData.Event.Description != "" {
		config.outputFunc(fmt.Sprintf("Info: %s", config.eventData.Event.Description))
	}
	config.outputFunc(fmt.Sprintf("Track: %s", config.eventData.Track.Name))
	config.outputFunc(fmt.Sprintf("Date: %s", config.eventData.Event.EventTime.AsTime()))
	config.outputFunc(fmt.Sprintf("Cars: %d", len(config.eventData.Car.CurrentDrivers)))
}

func (e *eventText) eventCars(config *EventOutputConfig) {
	for _, ce := range config.sortedEntries() {
		config.outputFunc(fmt.Sprintf("Car: %3s %s ", ce.Car.CarNumber, ce.Team.Name))
	}
}

func (e *eventText) eventCarLaps(config *EventOutputConfig) {
	extraInfo := func(l *analysisv1.Lap) string {
		if l.LapInfo == nil {
			return ""
		}
		return fmt.Sprintf("%s %.3f", l.LapInfo.Mode, l.LapInfo.Time)
	}
	for _, v := range config.eventData.Analysis.CarLaps {
		if config.showCar(v.CarNum) {
			config.outputFunc(fmt.Sprintf("Car #%s laps:", v.CarNum))
			for _, l := range v.Laps {
				config.outputFunc(fmt.Sprintf("  %d %.3f %s", l.LapNo, l.LapTime, extraInfo(l)))
			}
		}
	}
}
