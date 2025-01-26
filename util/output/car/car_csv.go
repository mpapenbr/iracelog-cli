package car

import (
	"encoding/csv"
	"fmt"

	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"
)

type carCsv struct {
	config *OutputConfig
	writer *csv.Writer
}

func newCarCsv(cfg *OutputConfig) *carCsv {
	return &carCsv{
		config: cfg,
		writer: csv.NewWriter(cfg.writer),
	}
}

func (c *carCsv) header() {
	headers := make([]string, len(c.config.attrs))
	for i, attr := range c.config.attrs {
		headers[i] = attr.String()
	}
	//nolint:errcheck // by design
	c.writer.Write(headers)
	c.writer.Flush()
}

func (c *carCsv) line(session *racestatev1.Session, car *racestatev1.Car) {
	record := make([]string, len(c.config.attrs))
	for i, attr := range c.config.attrs {
		record[i] = getCarAttrValue(session, car, attr)
	}
	//nolint:errcheck // by design
	c.writer.Write(record)
	c.writer.Flush()
}

func (c *carCsv) flush() {
	c.writer.Flush()
}

//nolint:exhaustive,funlen,gocyclo,whitespace // by design
func getCarAttrValue(
	session *racestatev1.Session,
	car *racestatev1.Car,
	attr CarAttr,
) string {
	switch attr {
	case CarAttrIdx:
		return fmt.Sprintf("%d", car.GetCarIdx())
	case CarAttrState:
		return car.GetState().String()
	case CarAttrPos:
		return fmt.Sprintf("%d", car.GetPos())
	case CarAttrPic:
		return fmt.Sprintf("%d", car.GetPic())
	case CarAttrLap:
		return fmt.Sprintf("%d", car.GetLap())
	case CarAttrLc:
		return fmt.Sprintf("%d", car.GetLc())
	case CarAttrTrackPos:
		return fmt.Sprintf("%.3f", car.GetTrackPos())
	case CarAttrPitstops:
		return fmt.Sprintf("%d", car.GetPitstops())
	case CarAttrStintLap:
		return fmt.Sprintf("%d", car.GetStintLap())
	case CarAttrSpeed:
		return fmt.Sprintf("%.0f", car.GetSpeed())
	case CarAttrDist:
		return fmt.Sprintf("%.0f", car.GetDist())
	case CarAttrInterval:
		return fmt.Sprintf("%.1f", car.GetInterval())
	case CarAttrGap:
		return fmt.Sprintf("%.1f", car.GetGap())
	case CarAttrTireCompound:
		return fmt.Sprintf("%d", car.GetTireCompound().GetRawValue())
	case CarSessionTime:
		return fmt.Sprintf("%.0f", session.SessionTime)
		// case CarAttrBestLap:
		// 	return fmt.Sprintf("%.3f", car.GetBest())
		// case CarAttrLastLap:
		// 	return fmt.Sprintf("%.3f", car.GetLast())
	}
	return ""
}
