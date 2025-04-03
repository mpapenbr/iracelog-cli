package session

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
		attrs      []SessionAttr
		outputFunc func(s string)
		writer     io.Writer
	}
	Output interface {
		Header()
		Line(data *racestatev1.PublishStateRequest)
		Flush()
	}

	sessionOutput struct {
		outputter formatOutput
	}
	formatOutput interface {
		header()
		line(data *racestatev1.PublishStateRequest)
		flush()
	}
)

func NewSessionOutput(opts ...Option) Output {
	cfg := &OutputConfig{
		outputFunc: func(s string) { fmt.Println(s) },
		format:     output.FormatText,
		attrs:      []SessionAttr{},
		writer:     os.Stdout,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	//nolint:exhaustive // by design
	switch cfg.format {
	case output.FormatCSV:
		return &sessionOutput{outputter: newSessionCsv(cfg)}
	case output.FormatJSON:
		return &sessionOutput{outputter: &sessionJSON{config: cfg}}
	}
	return &sessionOutput{outputter: &sessionEmpty{config: cfg}}
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

func WithSessionAttrs(attrs []SessionAttr) Option {
	return func(cfg *OutputConfig) {
		cfg.attrs = attrs
	}
}

func WithAllSessionAttrs() Option {
	return func(cfg *OutputConfig) {
		cfg.attrs = SupportedSessionAttrs()
	}
}

func (s *sessionOutput) Header() {
	s.outputter.header()
}

func (s *sessionOutput) Line(data *racestatev1.PublishStateRequest) {
	s.outputter.line(data)
}

func (s *sessionOutput) Flush() {
	s.outputter.flush()
}
