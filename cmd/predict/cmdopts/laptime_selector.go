package cmdopts

import (
	"bytes"
	"fmt"
	"strings"

	predictv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/predict/v1"
)

// laptimeSelector is a local type that embeds predictv1.laptimeSelector
type laptimeSelector struct {
	predictv1.LaptimeSelector
}

func ParseLaptimeSelector(text string) (predictv1.LaptimeSelector, error) {
	var c laptimeSelector
	err := c.UnmarshalText([]byte(text))
	return c.LaptimeSelector, err
}

func (c laptimeSelector) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c *laptimeSelector) UnmarshalText(text []byte) error {
	if c == nil {
		return fmt.Errorf("cannot unmarshal into nil LaptimeSelector")
	}
	if !c.unmarshalText(text) && !c.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized format: %q", text)
	}
	return nil
}

func (c *laptimeSelector) unmarshalText(text []byte) bool {
	switch strings.ToLower(string(text)) {
	case "prev-stint-avg", "": // make the zero value useful
		c.LaptimeSelector = predictv1.LaptimeSelector_LAPTIME_SELECTOR_PREVIOUS_STINT_AVG
	case "last":
		c.LaptimeSelector = predictv1.LaptimeSelector_LAPTIME_SELECTOR_LAST
	case "cur-stint-avg":
		c.LaptimeSelector = predictv1.LaptimeSelector_LAPTIME_SELECTOR_CURRENT_STINT_AVG
	default:
		return false
	}
	return true
}
