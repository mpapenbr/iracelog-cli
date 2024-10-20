package session

import racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"

type (
	sessionEmpty struct {
		config *OutputConfig
	}
)

func (s *sessionEmpty) header() {
	s.config.outputFunc("session header not implemented")
}

func (s *sessionEmpty) line(data *racestatev1.PublishStateRequest) {
	s.config.outputFunc("session line not implemented")
}

func (s *sessionEmpty) flush() {
	// empty by design
}
