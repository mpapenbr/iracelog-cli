package util

import (
	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
)

type (
	TenantSelParam interface {
		ExternalId() string
		Name() string
	}
)

func ResolveTenant(tsl TenantSelParam) *commonv1.TenantSelector {
	if tsl.ExternalId() != "" {
		return &commonv1.TenantSelector{
			Arg: &commonv1.TenantSelector_ExternalId{
				ExternalId: &commonv1.UUID{Id: tsl.ExternalId()},
			},
		}
	}
	if tsl.Name() != "" {
		return &commonv1.TenantSelector{
			Arg: &commonv1.TenantSelector_Name{Name: tsl.Name()},
		}
	}
	return nil
}
