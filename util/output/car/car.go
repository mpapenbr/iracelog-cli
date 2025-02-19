package car

import (
	"fmt"
	"io"
	"os"

	racestatev1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/racestate/v1"

	"github.com/mpapenbr/iracelog-cli/util/output"
)

type (
	Option       func(*OutputConfig)
	OutputConfig struct {
		format     output.Format
		attrs      []CarAttr
		outputFunc func(s string)
		writer     io.Writer
	}
	Output interface {
		Header()
		Line(session *racestatev1.Session, car *racestatev1.Car)
		Flush()
	}

	carOutput struct {
		outputter formatOutput
	}
	formatOutput interface {
		header()
		line(session *racestatev1.Session, car *racestatev1.Car)
		flush()
	}
)

func NewCarOutput(opts ...Option) Output {
	cfg := &OutputConfig{
		outputFunc: func(s string) { fmt.Println(s) },
		format:     output.FormatText,
		attrs:      []CarAttr{},
		writer:     os.Stdout,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	//nolint:exhaustive // by design
	switch cfg.format {
	case output.FormatCSV:
		return &carOutput{outputter: newCarCsv(cfg)}
	case output.FormatJSON:
		return &carOutput{outputter: &carJson{config: cfg}}
	}
	return &carOutput{outputter: &carEmpty{config: cfg}}
}

func WithFormat(f output.Format) Option {
	return func(cfg *OutputConfig) {
		cfg.format = f
	}
}

func WithWriter(w io.Writer) Option {
	return func(cfg *OutputConfig) {
		cfg.writer = w
	}
}

func WithCarAttrs(attrs []CarAttr) Option {
	return func(cfg *OutputConfig) {
		cfg.attrs = attrs
	}
}

func WithAllCarAttrs() Option {
	return func(cfg *OutputConfig) {
		cfg.attrs = SupportedCarAttrs()
	}
}

func (s *carOutput) Header() {
	s.outputter.header()
}

func (s *carOutput) Line(session *racestatev1.Session, car *racestatev1.Car) {
	s.outputter.line(session, car)
}

func (s *carOutput) Flush() {
	s.outputter.flush()
}
