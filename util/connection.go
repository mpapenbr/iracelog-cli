package util

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mpapenbr/iracelog-cli/config"
)

func ConnectGrpc(cfg *config.CliArgs) (*grpc.ClientConn, error) {
	return ConnectGrpcWithParam(cfg.Addr, cfg.Insecure, cfg.InsecureSkipVerify)
}

//nolint:whitespace // by design
func ConnectGrpcWithParam(addr string, noTls, insecureSkipVerify bool) (
	*grpc.ClientConn, error,
) {
	if noTls {
		return grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {

		// serverCert, err := os.ReadFile("cert.pem")
		// if err != nil {
		// 	log.Fatal("failed to read server certificate", log.ErrorField(err))
		// }
		// Create a certificate pool and add the server's certificate
		// caCertPool := x509.NewCertPool()
		// if ok := caCertPool.AppendCertsFromPEM(serverCert); !ok {
		// 	log.Fatal("failed to append server certificate")
		// }

		//nolint:gosec // by design
		tlsConfig := &tls.Config{
			MinVersion:         tls.VersionTLS13, // Set the minimum TLS version to TLS 1.3
			InsecureSkipVerify: insecureSkipVerify,
			// RootCAs:            caCertPool,
		}

		return grpc.NewClient(addr,
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}
}
