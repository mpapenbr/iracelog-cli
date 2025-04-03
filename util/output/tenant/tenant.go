package tenant

import (
	"fmt"
	"io"
	"os"

	tenantv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/tenant/v1"

	"github.com/mpapenbr/iracelog-cli/util/output"
)

type (
	Option       func(*OutputConfig)
	OutputConfig struct {
		format     output.Format
		attrs      []TenantAttr
		outputFunc func(s string)
		writer     io.Writer
	}
	Output interface {
		Header()
		Line(data *tenantv1.Tenant)
		Flush()
	}

	tenantOutput struct {
		outputter formatOutput
	}
	formatOutput interface {
		header()
		line(data *tenantv1.Tenant)
		flush()
	}
)

func NewTenantOutput(opts ...Option) Output {
	cfg := &OutputConfig{
		outputFunc: func(s string) { fmt.Println(s) },
		format:     output.FormatText,
		attrs:      []TenantAttr{},
		writer:     os.Stdout,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	switch cfg.format {
	case output.FormatCSV:
		return &tenantOutput{outputter: newTenantCsv(cfg)}
	case output.FormatJSON:
		return &tenantOutput{outputter: &tenantJSON{config: cfg}}
	case output.FormatText:
		return &tenantOutput{outputter: &tenantText{config: cfg}}
	}
	return &tenantOutput{outputter: &tenantEmpty{config: cfg}}
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

func WithTenantAttrs(attrs []TenantAttr) Option {
	return func(cfg *OutputConfig) {
		cfg.attrs = attrs
	}
}

func WithAllTenantAttrs() Option {
	return func(cfg *OutputConfig) {
		cfg.attrs = SupportedTenantAttrs()
	}
}

func (s *tenantOutput) Header() {
	s.outputter.header()
}

func (s *tenantOutput) Line(data *tenantv1.Tenant) {
	s.outputter.line(data)
}

func (s *tenantOutput) Flush() {
	s.outputter.flush()
}
