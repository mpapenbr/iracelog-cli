package tenant

import (
	tenantv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/tenant/v1"
)

type (
	tenantEmpty struct {
		config *OutputConfig
	}
)

func (s *tenantEmpty) header() {
	s.config.outputFunc("tenant header not implemented")
}

func (s *tenantEmpty) line(data *tenantv1.Tenant) {
	s.config.outputFunc("tenant line not implemented")
}

func (s *tenantEmpty) flush() {
	// empty by design
}
