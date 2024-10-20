package event

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mpapenbr/iracelog-cli/util/output"
)

type (
	Component int8
)

const (
	ComponentUnknown Component = iota
	ComponentEventInfo
	ComponentEventCars
	ComponentEventCarLaps
)

func ParseComponent(text string) (Component, error) {
	var c Component
	err := c.UnmarshalText([]byte(text))
	return c, err
}

func (c Component) String() string {
	switch c {
	case ComponentEventInfo:
		return "info"
	case ComponentEventCars:
		return "cars"
	case ComponentEventCarLaps:
		return "carlaps"
	case ComponentUnknown:
		return output.Unknown
	default:
		return output.Unknown
	}
}

func (c Component) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c *Component) UnmarshalText(text []byte) error {
	if c == nil {
		return output.ErrUnmarshalNil
	}
	if !c.unmarshalText(text) && !c.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized format: %q", text)
	}
	return nil
}

func (c *Component) unmarshalText(text []byte) bool {
	switch strings.ToLower(string(text)) {
	case "info", "": // make the zero value useful
		*c = ComponentEventInfo
	case "cars":
		*c = ComponentEventCars
	case "carlaps":
		*c = ComponentEventCarLaps
	default:
		return false
	}
	return true
}
