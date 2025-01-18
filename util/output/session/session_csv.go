package session

import (
	"encoding/csv"
	"fmt"

	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
)

type (
	sessionCsv struct {
		config *OutputConfig
		writer *csv.Writer
	}
)

func newSessionCsv(config *OutputConfig) *sessionCsv {
	ret := &sessionCsv{
		config: config,
		writer: csv.NewWriter(config.writer),
	}
	return ret
}

func (s *sessionCsv) header() {
	data := []string{}
	for _, attr := range s.config.attrs {
		data = append(data, attr.String())
	}
	//nolint:errcheck // by design
	s.writer.Write(data)
}

//nolint:cyclop,funlen // by design
func (s *sessionCsv) line(data *racestatev1.PublishStateRequest) {
	out := []string{}
	for _, attr := range s.config.attrs {
		var valueString string
		switch attr {
		case SessionTime:
			valueString = fmt.Sprintf("%.3f", data.Session.GetSessionTime())
		case SessionNum:
			valueString = fmt.Sprintf("%d", data.Session.GetSessionNum())
		case SessionTimeOfDay:
			valueString = fmt.Sprintf("%d", data.Session.GetTimeOfDay())
		case SessionLapsRemain:
			valueString = fmt.Sprintf("%d", data.Session.GetLapsRemain())
		case SessionTimeRemain:
			valueString = fmt.Sprintf("%.3f", data.Session.GetTimeRemain())
		case SessionTrackTemp:
			valueString = fmt.Sprintf("%.2f", data.Session.GetTrackTemp())
		case SessionAirTemp:
			valueString = fmt.Sprintf("%.2f", data.Session.GetTrackTemp())
		case SessionTrackWetness:
			valueString = data.Session.GetTrackWetness().String()
		case SessionPrecipitation:
			valueString = fmt.Sprintf("%.4f", data.Session.GetPrecipitation())
		case SessionFlagState:
			valueString = data.Session.GetFlagState()
		case SessionTimestamp:
			valueString = data.Timestamp.AsTime().String()
		case SessionUndefined:
			valueString = "undefined"
		default:
			valueString = "unknown"
		}
		out = append(out, valueString)
	}
	//nolint:errcheck // by design
	s.writer.Write(out)
}

func (s *sessionCsv) flush() {
	s.writer.Flush()
}
