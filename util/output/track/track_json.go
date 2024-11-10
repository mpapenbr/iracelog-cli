package track

import (
	"encoding/json"

	trackv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/track/v1"
)

type (
	trackJson struct {
		config *OutputConfig
	}
)

func (s *trackJson) header() {
	// empty by design - not needed for json
}

//nolint:cyclop // by design
func (s *trackJson) line(data *trackv1.Track) {
	out := make(map[string]interface{}, 0)
	for _, attr := range s.config.attrs {
		//nolint:exhaustive // by design
		switch attr {
		case TrackId:
			out[attr.String()] = data.Id
		case TrackName:
			out[attr.String()] = data.Name
		case TrackShortName:
			out[attr.String()] = data.ShortName
		case TrackConfig:
			out[attr.String()] = data.Config
		case TrackLength:
			out[attr.String()] = data.Length
		case TrackPitSpeed:
			out[attr.String()] = data.PitSpeed
		case TrackPitEntry:
			out[attr.String()] = data.PitInfo.Entry
		case TrackPitExit:
			out[attr.String()] = data.PitInfo.Exit
		case TrackPitLaneLength:
			out[attr.String()] = data.PitInfo.LaneLength

		case TrackUndefined:
			// empty by design
		}
	}
	//nolint:errcheck // by design
	if jsonData, err := json.Marshal(out); err == nil {
		s.config.writer.Write(jsonData)
		s.config.writer.Write([]byte("\n"))
	}
}

func (s *trackJson) flush() {
	// empty by design - not needed for json
}
