package car

import (
	"encoding/json"
	"fmt"

	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
)

type carJSON struct {
	config *OutputConfig
}

func (c *carJSON) header() {
	// JSON output does not require a header
}

func (c *carJSON) line(session *racestatev1.Session, car *racestatev1.Car) {
	record := make(map[string]interface{})
	for _, attr := range c.config.attrs {
		record[attr.String()] = getCarAttrValue(session, car, attr)
	}
	jsonData, err := json.Marshal(record)
	if err != nil {
		c.config.outputFunc(fmt.Sprintf("error marshaling JSON: %v", err))
		return
	}
	c.config.outputFunc(string(jsonData))
}

func (c *carJSON) flush() {
	// JSON output does not require flushing
}
