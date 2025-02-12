package util

import (
	"time"

	commonv1 "buf.build/gen/go/mpapenbr/iracelog/protocolbuffers/go/iracelog/common/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type (
	StartSelParam interface {
		SessionTime() time.Duration
		RecordStamp() string
		SessionNum() int
		Id() int
	}
)

//nolint:whitespace // can't make all linters happy
func ResolveStartSelector(sessionTime time.Duration, recordStamp string) (
	*commonv1.StartSelector, error,
) {
	if sessionTime > 0 {
		return &commonv1.StartSelector{
			Arg: &commonv1.StartSelector_SessionTime{
				SessionTime: float32(sessionTime.Seconds()),
			},
		}, nil
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

//nolint:whitespace // can't make all linters happy
func ResolveStartSelector2(s StartSelParam) (
	*commonv1.StartSelector, error,
) {
	if s.RecordStamp() != "" {
		if val, err := time.Parse(time.RFC3339, s.RecordStamp()); err != nil {
			return nil, err
		} else {
			return &commonv1.StartSelector{
				Arg: &commonv1.StartSelector_RecordStamp{
					RecordStamp: timestamppb.New(val),
				},
			}, nil
		}
	}
	if s.SessionTime() > -1 {
		if s.SessionNum() > -1 {
			return &commonv1.StartSelector{
				Arg: &commonv1.StartSelector_SessionTimeSelector{
					SessionTimeSelector: &commonv1.SessionTimeSelector{
						Num:      int32(s.SessionNum()),
						Duration: durationpb.New(s.SessionTime()),
					},
				},
			}, nil
		}
		return &commonv1.StartSelector{
			Arg: &commonv1.StartSelector_SessionTime{
				SessionTime: float32(s.SessionTime().Seconds()),
			},
		}, nil
	}

	if s.Id() > -1 {
		return &commonv1.StartSelector{
			Arg: &commonv1.StartSelector_Id{
				Id: uint32(s.Id()),
			},
		}, nil
	}
	return nil, nil
}
