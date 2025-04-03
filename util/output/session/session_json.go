package session

import (
	"encoding/json"

	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
)

type (
	sessionJSON struct {
		config *OutputConfig
	}
)

func (s *sessionJSON) header() {
	// empty by design - not needed for json
}

//nolint:cyclop // by design
func (s *sessionJSON) line(data *racestatev1.PublishStateRequest) {
	out := make(map[string]interface{}, 0)
	for _, attr := range s.config.attrs {
		switch attr {
		case SessionTime:
			out[attr.String()] = data.Session.GetSessionTime()
		case SessionNum:
			out[attr.String()] = data.Session.GetSessionNum()
		case SessionTimeOfDay:
			out[attr.String()] = data.Session.GetTimeOfDay()
		case SessionLapsRemain:
			out[attr.String()] = data.Session.GetLapsRemain()
		case SessionTimeRemain:
			out[attr.String()] = data.Session.GetTimeRemain()
		case SessionTrackTemp:
			out[attr.String()] = data.Session.GetTrackTemp()
		case SessionAirTemp:
			out[attr.String()] = data.Session.GetTrackTemp()
		case SessionTrackWetness:
			out[attr.String()] = data.Session.GetTrackWetness().String()
		case SessionPrecipitation:
			out[attr.String()] = data.Session.GetPrecipitation()
		case SessionFlagState:
			out[attr.String()] = data.Session.GetFlagState()
		case SessionTimestamp:
			out[attr.String()] = data.Timestamp.AsTime().String()
		case SessionUndefined:
			// empty by design
		}
	}
	//nolint:errcheck // by design
	if jsonData, err := json.Marshal(out); err == nil {
		s.config.writer.Write(jsonData)
		s.config.writer.Write([]byte("\n"))
	}
}

func (s *sessionJSON) flush() {
	// empty by design - not needed for json
}
