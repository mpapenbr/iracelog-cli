package cmdopts

import (
	"github.com/mpapenbr/iracelog-cli/util/output"
	"github.com/mpapenbr/iracelog-cli/util/output/tenant"
)

func ConfigureOutput(format string, attrs []string) tenant.Output {
	opts := []tenant.Option{}

	if format != "" {
		if f, err := output.ParseFormat(format); err == nil {
			opts = append(opts, tenant.WithFormat(f))
		}
	}
	if len(attrs) > 0 {
		tenantAttrs := []tenant.TenantAttr{}
		for _, c := range attrs {
			v, _ := tenant.ParseTenantAttr(c)
			tenantAttrs = append(tenantAttrs, v)
		}
		opts = append(opts, tenant.WithTenantAttrs(tenantAttrs))
	} else {
		opts = append(opts, tenant.WithAllTenantAttrs())
	}
	return tenant.NewTenantOutput(opts...)
}
