package output

import (
	"bytes"
	"errors"
	"fmt"
)

type (
	Format int8
)

const (
	FormatText Format = iota
	FormatJSON
	FormatCSV
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
	case FormatCSV:
		return "csv"
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
	case "csv", "CSV":
		*f = FormatCSV
	case "text", "TEXT", "": // make the zero value useful
		*f = FormatText
	default:
		return false
	}
	return true
}
