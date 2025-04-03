package track

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mpapenbr/iracelog-cli/util/output"
)

const (
	TrackUndefined TrackAttr = iota
	TrackID
	TrackName
	TrackShortName
	TrackConfig
	TrackLength
	TrackSectors
	TrackPitSpeed
	TrackPitEntry
	TrackPitExit
	TrackPitLaneLength
)

type (
	TrackAttr int8
)

func SupportedTrackAttrs() []TrackAttr {
	return []TrackAttr{
		TrackID,
		TrackName,
		TrackShortName,
		TrackConfig,
		TrackLength,
		// TrackSectors, // not supported atm
		TrackPitSpeed,
		TrackPitEntry,
		TrackPitExit,
		TrackPitLaneLength,
	}
}

func ParseTrackAttr(text string) (TrackAttr, error) {
	var f TrackAttr
	err := f.UnmarshalText([]byte(text))
	return f, err
}

//nolint:exhaustive,cyclop // by design
func (f TrackAttr) String() string {
	switch f {
	case TrackID:
		return "id"
	case TrackName:
		return "name"
	case TrackShortName:
		return "shortName"
	case TrackConfig:
		return "config"
	case TrackLength:
		return "length"
	case TrackSectors:
		return "sectors"
	case TrackPitSpeed:
		return "pitSpeed"
	case TrackPitEntry:
		return "pitEntry"
	case TrackPitExit:
		return "pitExit"
	case TrackPitLaneLength:
		return "pitLaneLength"

	default:
		return output.Unknown
	}
}

func (f TrackAttr) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func (f *TrackAttr) UnmarshalText(text []byte) error {
	if f == nil {
		return output.ErrUnmarshalNil
	}
	if !f.unmarshalText(text) && !f.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized session attr: %q", text)
	}
	return nil
}

//nolint:cyclop // by design
func (f *TrackAttr) unmarshalText(text []byte) bool {
	switch strings.ToLower(string(text)) {
	case "id":
		*f = TrackID
	case "name":
		*f = TrackName
	case "shortname":
		*f = TrackShortName
	case "config":
		*f = TrackConfig
	case "length":
		*f = TrackLength
	case "sectors":
		*f = TrackSectors
	case "pitspeed":
		*f = TrackPitSpeed
	case "pitentry":
		*f = TrackPitEntry
	case "pitexit":
		*f = TrackPitExit
	case "pitlanelength":
		*f = TrackPitLaneLength
	default:
		return false
	}
	return true
}
