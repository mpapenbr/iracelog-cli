package util

import (
	"strconv"

	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
)

//nolint:gosec //check is not needed here
func ResolveEvent(arg string) *commonv1.EventSelector {
	if id, err := strconv.Atoi(arg); err == nil {
		return &commonv1.EventSelector{Arg: &commonv1.EventSelector_Id{Id: int32(id)}}
	}
	return &commonv1.EventSelector{Arg: &commonv1.EventSelector_Key{Key: arg}}
}
