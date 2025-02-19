package tenant

import (
	"fmt"
	"strings"

	tenantv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/tenant/v1"
)

type (
	tenantText struct {
		config *OutputConfig
	}
)

func (s *tenantText) header() {
	// empty by design
}

//nolint:cyclop,errcheck // by design
func (s *tenantText) line(data *tenantv1.Tenant) {
	out := []string{}
	for _, attr := range s.config.attrs {
		var valueString string
		switch attr {
		case TenantId:
			valueString = fmt.Sprintf("%.d", data.GetId())
		case TenantExternalId:
			valueString = data.GetExternalId().GetId()
		case TenantName:
			valueString = data.GetName()
		case TenantApiKey:
			valueString = data.GetApiKey()
		case TenantIsActive:
			valueString = fmt.Sprintf("%t", data.GetIsActive())
		case TenantUndefined:
			valueString = "undefined"
		default:
			valueString = "unknown"
		}
		out = append(out, fmt.Sprintf("%s=%s", attr, valueString))
	}
	s.config.writer.Write([]byte(strings.Join(out, " ")))
	s.config.writer.Write([]byte("\n"))
}

func (s *tenantText) flush() {
	// empty by design
}
