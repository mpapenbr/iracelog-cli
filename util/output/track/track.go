package track

import (
	"fmt"
	"io"
	"os"

	trackv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/track/v1"

	"github.com/mpapenbr/iracelog-cli/util/output"
)

type (
	Option       func(*OutputConfig)
	OutputConfig struct {
		format     output.Format
		attrs      []TrackAttr
		outputFunc func(s string)
		writer     io.Writer
	}
	Output interface {
		Header()
		Line(data *trackv1.Track)
		Flush()
	}

	trackOutput struct {
		outputter formatOutput
	}
	formatOutput interface {
		header()
		line(data *trackv1.Track)
		flush()
	}
)

func NewTrackOutput(opts ...Option) Output {
	cfg := &OutputConfig{
		outputFunc: func(s string) { fmt.Println(s) },
		format:     output.FormatText,
		attrs:      []TrackAttr{},
		writer:     os.Stdout,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	//nolint:exhaustive // by design
	switch cfg.format {
	case output.FormatCSV:
		return &trackOutput{outputter: newTrackCsv(cfg)}
	case output.FormatJSON:
		return &trackOutput{outputter: &trackJSON{config: cfg}}
	}
	return &trackOutput{outputter: &trackEmpty{config: cfg}}
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

func WithTrackAttrs(attrs []TrackAttr) Option {
	return func(cfg *OutputConfig) {
		cfg.attrs = attrs
	}
}

func WithAllTrackAttrs() Option {
	return func(cfg *OutputConfig) {
		cfg.attrs = SupportedTrackAttrs()
	}
}

func (s *trackOutput) Header() {
	s.outputter.header()
}

func (s *trackOutput) Line(data *trackv1.Track) {
	s.outputter.line(data)
}

func (s *trackOutput) Flush() {
	s.outputter.flush()
}
