package util

import (
	"time"

	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//nolint:whitespace // can't make all linters happy
func ResolveStartSelector(sessionTime, recordStamp string) (
	*commonv1.StartSelector, error,
) {
	if sessionTime != "" {
		if val, err := time.ParseDuration(sessionTime); err != nil {
			return nil, err
		} else {
			return &commonv1.StartSelector{
				Arg: &commonv1.StartSelector_SessionTime{
					SessionTime: float32(val.Seconds()),
				},
			}, nil
		}
	}
	if recordStamp != "" {
		if val, err := time.Parse(time.RFC3339, recordStamp); err != nil {
			return nil, err
		} else {
			return &commonv1.StartSelector{
				Arg: &commonv1.StartSelector_RecordStamp{
					RecordStamp: timestamppb.New(val),
				},
			}, nil
		}
	}
	return nil, nil
}
