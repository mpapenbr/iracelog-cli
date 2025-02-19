package tenant

import (
	"encoding/csv"
	"fmt"

	tenantv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/tenant/v1"
)

type (
	tenantCsv struct {
		config *OutputConfig
		writer *csv.Writer
	}
)

func newTenantCsv(config *OutputConfig) *tenantCsv {
	ret := &tenantCsv{
		config: config,
		writer: csv.NewWriter(config.writer),
	}
	return ret
}

func (s *tenantCsv) header() {
	data := []string{}
	for _, attr := range s.config.attrs {
		data = append(data, attr.String())
	}
	//nolint:errcheck // by design
	s.writer.Write(data)
}

//nolint:cyclop // by design
func (s *tenantCsv) line(data *tenantv1.Tenant) {
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
		out = append(out, valueString)
	}
	//nolint:errcheck // by design
	s.writer.Write(out)
}

func (s *tenantCsv) flush() {
	s.writer.Flush()
}
