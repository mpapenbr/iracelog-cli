package tenant

import (
	"encoding/json"

	tenantv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/tenant/v1"
)

type (
	tenantJSON struct {
		config *OutputConfig
	}
)

func (s *tenantJSON) header() {
	// empty by design - not needed for json
}

//nolint:cyclop // by design
func (s *tenantJSON) line(data *tenantv1.Tenant) {
	out := make(map[string]interface{}, 0)
	for _, attr := range s.config.attrs {
		switch attr {
		case TenantExternalID:
			out[attr.String()] = data.GetExternalId().GetId()
		case TenantName:
			out[attr.String()] = data.GetName()
		case TenantIsActive:
			out[attr.String()] = data.GetIsActive()

		case TenantUndefined:
			// empty by design
		}
	}
	//nolint:errcheck // by design
	if jsonData, err := json.Marshal(out); err == nil {
		s.config.writer.Write(jsonData)
		s.config.writer.Write([]byte("\n"))
	}
}

func (s *tenantJSON) flush() {
	// empty by design - not needed for json
}
