package track

import (
	"encoding/csv"
	"fmt"

	trackv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/track/v1"
)

type (
	trackCsv struct {
		config *OutputConfig
		writer *csv.Writer
	}
)

func newTrackCsv(config *OutputConfig) *trackCsv {
	ret := &trackCsv{
		config: config,
		writer: csv.NewWriter(config.writer),
	}
	return ret
}

func (s *trackCsv) header() {
	data := []string{}
	for _, attr := range s.config.attrs {
		data = append(data, attr.String())
	}
	//nolint:errcheck // by design
	s.writer.Write(data)
}

//nolint:cyclop // by design
func (s *trackCsv) line(data *trackv1.Track) {
	out := []string{}
	for _, attr := range s.config.attrs {
		var valueString string
		//nolint:exhaustive // by design
		switch attr {
		case TrackID:
			valueString = fmt.Sprintf("%d", data.Id)
		case TrackName:
			valueString = data.Name
		case TrackShortName:
			valueString = data.ShortName
		case TrackConfig:
			valueString = data.Config
		case TrackLength:
			valueString = fmt.Sprintf("%.3f", data.Length)
		case TrackPitSpeed:
			valueString = fmt.Sprintf("%.3f", data.PitSpeed)
		case TrackPitEntry:
			valueString = fmt.Sprintf("%.3f", data.PitInfo.Entry)
		case TrackPitExit:
			valueString = fmt.Sprintf("%.3f", data.PitInfo.Exit)
		case TrackPitLaneLength:
			valueString = fmt.Sprintf("%.3f", data.PitInfo.LaneLength)
		case TrackUndefined:
			valueString = "undefined"
		default:
			valueString = "unknown"
		}
		out = append(out, valueString)
	}
	//nolint:errcheck // by design
	s.writer.Write(out)
}

func (s *trackCsv) flush() {
	s.writer.Flush()
}
