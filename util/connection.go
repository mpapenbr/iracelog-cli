package util

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mpapenbr/iracelog-cli/config"
)

func ConnectGrpc(cfg *config.CliArgs) (*grpc.ClientConn, error) {
	return ConnectGrpcWithParam(cfg.Addr, cfg.Insecure)
}

func ConnectGrpcWithParam(addr string, noTls bool) (*grpc.ClientConn, error) {
	if noTls {
		return grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS13, // Set the minimum TLS version to TLS 1.3
		}
		return grpc.NewClient(addr,
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}
}
