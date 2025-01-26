package options

import (
	"time"
)

var (
	SessionTime time.Duration // session time where data should begin (for example: 10m)
	RecordStamp string        // timestamp time where data should begin
	NumEntries  int32         // number of entries to process

)
