package output

type (
	eventEmpty struct{}
)

func (e *eventEmpty) eventInfo(config *EventOutputConfig) {
	config.outputFunc("eventInfo not implemented")
}

func (e *eventEmpty) eventCars(config *EventOutputConfig) {
	config.outputFunc("eventCars not implemented")
}

func (e *eventEmpty) eventCarLaps(config *EventOutputConfig) {
	config.outputFunc("eventCarLaps not implemented")
}
