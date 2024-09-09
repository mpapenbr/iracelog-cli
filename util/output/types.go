package output

import (
	"bytes"
	"errors"
	"fmt"
)

type (
	Component int8
	Format    int8
)

const (
	ComponentUnknown Component = iota
	ComponentEventInfo
	ComponentEventCars
	ComponentEventCarLaps
)

const (
	FormatText Format = iota
	FormatJSON
)

const (
	Unknown = "unknown"
)

var ErrUnmarshalNil = errors.New("can't unmarshal a nil value")

func ParseFormat(text string) (Format, error) {
	var f Format
	err := f.UnmarshalText([]byte(text))
	return f, err
}

func (f Format) String() string {
	switch f {
	case FormatText:
		return "text"
	case FormatJSON:
		return "json"
	default:
		return Unknown
	}
}

func (f Format) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func (f *Format) UnmarshalText(text []byte) error {
	if f == nil {
		return ErrUnmarshalNil
	}
	if !f.unmarshalText(text) && !f.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized format: %q", text)
	}
	return nil
}

func (f *Format) unmarshalText(text []byte) bool {
	switch string(text) {
	case "json", "JSON":
		*f = FormatJSON
	case "text", "TEXT", "": // make the zero value useful
		*f = FormatText
	default:
		return false
	}
	return true
}

// Component starts here

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
		return Unknown
	default:
		return Unknown
	}
}

func (c Component) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c *Component) UnmarshalText(text []byte) error {
	if c == nil {
		return ErrUnmarshalNil
	}
	if !c.unmarshalText(text) && !c.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized format: %q", text)
	}
	return nil
}

func (c *Component) unmarshalText(text []byte) bool {
	switch string(text) {
	case "info", "INFO", "": // make the zero value useful
		*c = ComponentEventInfo
	case "cars", "CARS":
		*c = ComponentEventCars
	case "carlaps", "CARLAPS":
		*c = ComponentEventCarLaps
	default:
		return false
	}
	return true
}
