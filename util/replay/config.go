package replay

import "time"

type Config struct {
	Speed          int
	SourceAddr     string // grpc server address providing the data
	SourceInsecure bool   // connect to gRPC server without TLS
	Token          string
	EventKey       string
	DoNotPersist   bool
	FastForward    time.Duration
	FFPreRace      bool
}

func DefaultConfig() *Config {
	return &Config{
		Speed:          1,
		SourceAddr:     "",
		SourceInsecure: false,
		Token:          "",
		EventKey:       "",
		DoNotPersist:   false,
		FastForward:    time.Duration(0),
		FFPreRace:      true,
	}
}
