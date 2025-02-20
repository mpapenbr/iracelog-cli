package tenant

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/mpapenbr/iracelog-cli/util/output"
)

const (
	TenantUndefined TenantAttr = iota
	TenantExternalId
	TenantName
	TenantIsActive
)

type (
	TenantAttr int8
)

func ParseTenantAttr(text string) (TenantAttr, error) {
	var f TenantAttr
	err := f.UnmarshalText([]byte(text))
	return f, err
}

func SupportedTenantAttrs() []TenantAttr {
	return []TenantAttr{
		TenantExternalId,
		TenantName,
		TenantIsActive,
	}
}

//nolint:exhaustive,cyclop // by design
func (f TenantAttr) String() string {
	switch f {
	case TenantExternalId:
		return "externalId"
	case TenantName:
		return "name"
	case TenantIsActive:
		return "isactive"
	default:
		return output.Unknown
	}
}

func (f TenantAttr) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func (f *TenantAttr) UnmarshalText(text []byte) error {
	if f == nil {
		return output.ErrUnmarshalNil
	}
	if !f.unmarshalText(text) && !f.unmarshalText(bytes.ToLower(text)) {
		return fmt.Errorf("unrecognized session attr: %q", text)
	}
	return nil
}

//nolint:cyclop // by design
func (f *TenantAttr) unmarshalText(text []byte) bool {
	switch strings.ToLower(string(text)) {
	case "externalid":
		*f = TenantExternalId
	case "name":
		*f = TenantName
	case "isactive":
		*f = TenantIsActive

	default:
		return false
	}
	return true
}
