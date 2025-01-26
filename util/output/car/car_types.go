package car

import (
	"bytes"
	"fmt"
	"strings"
)

type (
	CarAttr int8
)

const (
	CarAttrIdx CarAttr = iota
	CarAttrState
	CarAttrPos
	CarAttrPic
	CarAttrLap
	CarAttrLc
	CarAttrTrackPos
	CarAttrPitstops
	CarAttrStintLap
	CarAttrSpeed
	CarAttrDist
	CarAttrInterval
	CarAttrGap
	CarAttrTireCompound
	CarAttrLastLap
	CarAttrBestLap
	CarSessionTime
)

func ParseCarAttr(text string) (CarAttr, error) {
	var f CarAttr
	err := f.UnmarshalText([]byte(text))
	return f, err
}

func SupportedCarAttrs() []CarAttr {
	return []CarAttr{
		CarAttrIdx,
		CarAttrState,
		CarAttrPos,
		CarAttrPic,
		CarAttrLap,
		CarAttrLc,
		CarAttrTrackPos,
		CarAttrPitstops,
		CarAttrStintLap,
		CarAttrSpeed,
		CarAttrDist,
		CarAttrInterval,
		CarAttrGap,
		CarAttrTireCompound,
		CarAttrLastLap,
		CarAttrBestLap,
		CarSessionTime,
	}
}

//nolint:gocyclo,funlen // by design
func (f CarAttr) String() string {
	switch f {
	case CarAttrIdx:
		return "idx"
	case CarAttrState:
		return "state"
	case CarAttrPos:
		return "pos"
	case CarAttrPic:
		return "pic"
	case CarAttrLap:
		return "lap"
	case CarAttrLc:
		return "lc"
	case CarAttrTrackPos:
		return "trackpos"
	case CarAttrPitstops:
		return "pitstops"
	case CarAttrStintLap:
		return "stintlap"
	case CarAttrSpeed:
		return "speed"
	case CarAttrDist:
		return "dist"
	case CarAttrInterval:
		return "interval"
	case CarAttrGap:
		return "gap"
	case CarAttrTireCompound:
		return "tirecompound"
	case CarAttrLastLap:
		return "lastlap"
	case CarAttrBestLap:
		return "bestlap"
	case CarSessionTime:
		return "sessiontime"
	default:
		return "unknown"
	}
}

func (f *CarAttr) UnmarshalText(text []byte) error {
	if f == nil {
		return fmt.Errorf("cannot unmarshal into nil CarAttr")
	}
	if !f.unmarshalText(text) && !f.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized format: %q", text)
	}
	return nil
}

//nolint:gocyclo,funlen // by design
func (f *CarAttr) unmarshalText(text []byte) bool {
	switch strings.ToLower(string(text)) {
	case "idx":
		*f = CarAttrIdx
	case "state":
		*f = CarAttrState
	case "pos":
		*f = CarAttrPos
	case "pic":
		*f = CarAttrPic
	case "lap":
		*f = CarAttrLap
	case "lc":
		*f = CarAttrLc
	case "trackpos":
		*f = CarAttrTrackPos
	case "pitstops":
		*f = CarAttrPitstops
	case "stintlap":
		*f = CarAttrStintLap
	case "speed":
		*f = CarAttrSpeed
	case "dist":
		*f = CarAttrDist
	case "interval":
		*f = CarAttrInterval
	case "gap":
		*f = CarAttrGap
	case "tirecompound":
		*f = CarAttrTireCompound
	case "lastlap":
		*f = CarAttrLastLap
	case "bestlap":
		*f = CarAttrBestLap
	case "sessiontime":
		*f = CarSessionTime
	default:
		return false
	}
	return true
}
