package session

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mpapenbr/iracelog-cli/util/output"
)

const (
	SessionUndefined SessionAttr = iota
	SessionTime
	SessionNum
	SessionTimeRemain
	SessionLapsRemain
	SessionTimeOfDay
	SessionTrackTemp
	SessionTrackWetness
	SessionPrecipitation
	SessionAirTemp
	SessionFlagState
)

type (
	SessionAttr int8
)

func ParseSessionAttr(text string) (SessionAttr, error) {
	var f SessionAttr
	err := f.UnmarshalText([]byte(text))
	return f, err
}

//nolint:exhaustive,cyclop // by design
func (f SessionAttr) String() string {
	switch f {
	case SessionTime:
		return "time"
	case SessionNum:
		return "num"
	case SessionTimeRemain:
		return "timeRemain"
	case SessionLapsRemain:
		return "lapsRemain"
	case SessionTimeOfDay:
		return "timeOfDay"
	case SessionTrackTemp:
		return "trackTemp"
	case SessionTrackWetness:
		return "trackWetness"
	case SessionPrecipitation:
		return "precipitation"
	case SessionAirTemp:
		return "airTemp"
	case SessionFlagState:
		return "flagState"

	default:
		return output.Unknown
	}
}

func (f SessionAttr) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func (f *SessionAttr) UnmarshalText(text []byte) error {
	if f == nil {
		return output.ErrUnmarshalNil
	}
	if !f.unmarshalText(text) && !f.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized session attr: %q", text)
	}
	return nil
}

//nolint:cyclop // by design
func (f *SessionAttr) unmarshalText(text []byte) bool {
	switch strings.ToLower(string(text)) {
	case "time":
		*f = SessionTime
	case "num":
		*f = SessionNum
	case "timeremain":
		*f = SessionTimeRemain
	case "lapsremain":
		*f = SessionLapsRemain
	case "timeofday":
		*f = SessionTimeOfDay
	case "tracktemp":
		*f = SessionTrackTemp
	case "trackwetness":
		*f = SessionTrackWetness
	case "precipitation":
		*f = SessionPrecipitation
	case "airtemp":
		*f = SessionAirTemp
	case "flagstate":
		*f = SessionFlagState
	default:
		return false
	}
	return true
}
