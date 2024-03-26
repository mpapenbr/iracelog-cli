package config

type CliArgs struct {
	Addr     string // ism gRPC address
	LogLevel string // sets the log level (zap log level values)
	Token    string // token for authentication
}

var cliArgs = NewCliArgs()

func DefaultCliArgs() *CliArgs {
	return cliArgs
}

func NewCliArgs() *CliArgs {
	return &CliArgs{}
}
