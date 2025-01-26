package car

import (
	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
)

type carEmpty struct {
	config *OutputConfig
}

func (c *carEmpty) header() {
	// No output
}

func (c *carEmpty) line(session *racestatev1.Session, car *racestatev1.Car) {
	// No output
}

func (c *carEmpty) flush() {
	// No output
}
