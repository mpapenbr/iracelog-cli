package options

import (
	"time"

	"github.com/mpapenbr/iracelog-cli/util"
)

var (
	SessionTime time.Duration // session time where data should begin (for example: 10m)
	SessionNum  int32         // Session num to be used
	RecordStamp string        // timestamp time where data should begin
	NumEntries  int32         // number of entries to process
	Id          int32         // Id to be used for start selector (internal sequence)

)

type myStartSelParam struct{}

func (m myStartSelParam) SessionTime() time.Duration {
	return SessionTime
}

func (m myStartSelParam) RecordStamp() string {
	return RecordStamp
}

func (m myStartSelParam) SessionNum() int {
	return int(SessionNum)
}

func (m myStartSelParam) Id() int {
	return int(Id)
}

func BuildStartSelParam() util.StartSelParam {
	return myStartSelParam{}
}
