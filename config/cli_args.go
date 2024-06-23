package config

import "fmt"

type CliArgs struct {
	Addr         string   // ism gRPC address
	Insecure     bool     // connect to gRPC server without TLS
	LogLevel     string   // sets the log level (zap log level values)
	LogFormat    string   // text vs json
	LogFile      string   // log file to write to
	Token        string   // token for authentication
	Event        string   // take event argument
	Components   []string // takes components for analysis selector
	DoNotPersist bool     // do not persist the recorded data (used for debugging)
}

func (c *CliArgs) Dump() {
	fmt.Printf("Addr: %s\n", c.Addr)
	fmt.Printf("LogLevel: %s\n", c.LogLevel)
	fmt.Printf("Token: %s\n", c.Token)
	fmt.Printf("Components: %v\n", c.Components)
}

var cliArgs = NewCliArgs()

func DefaultCliArgs() *CliArgs {
	return cliArgs
}

func NewCliArgs() *CliArgs {
	return &CliArgs{}
}
