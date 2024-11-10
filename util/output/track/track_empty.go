package track

import (
	trackv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/track/v1"
)

type (
	trackEmpty struct {
		config *OutputConfig
	}
)

func (s *trackEmpty) header() {
	s.config.outputFunc("session header not implemented")
}

func (s *trackEmpty) line(data *trackv1.Track) {
	s.config.outputFunc("session line not implemented")
}

func (s *trackEmpty) flush() {
	// empty by design
}
