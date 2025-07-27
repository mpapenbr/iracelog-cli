package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mpapenbr/iracelog-cli/config"
	"github.com/mpapenbr/iracelog-cli/log"
	"github.com/mpapenbr/iracelog-cli/util/cookie"
)

type (
	Option func(*param)
	param  struct {
		addr       string
		tlsEnabled bool
		cert       string
		key        string
		ca         string
		skipVerify bool
	}
)

func WithAddr(a string) Option {
	return func(p *param) {
		p.addr = a
	}
}

func WithTLSEnabled(b bool) Option {
	return func(p *param) {
		p.tlsEnabled = b
	}
}

func WithCert(filename string) Option {
	return func(p *param) {
		p.cert = filename
	}
}

func WithKey(filename string) Option {
	return func(p *param) {
		p.key = filename
	}
}

func WithCA(filename string) Option {
	return func(p *param) {
		p.ca = filename
	}
}

func WithSkipVerify(b bool) Option {
	return func(p *param) {
		p.skipVerify = b
	}
}

func WithCliArgs(args *config.CliArgs) Option {
	return func(p *param) {
		if args.Addr != "" {
			p.addr = args.Addr
		}
		if args.Insecure {
			p.tlsEnabled = false
		}
		if args.TLSCert != "" {
			p.cert = args.TLSCert
		}
		if args.TLSKey != "" {
			p.key = args.TLSKey
		}
		if args.TLSCa != "" {
			p.ca = args.TLSCa
		}

		p.skipVerify = args.TLSSkipVerify
	}
}

func NewClient(addr string, opts ...Option) (*grpc.ClientConn, error) {
	param := &param{addr: addr, tlsEnabled: true}
	for _, opt := range opts {
		opt(param)
	}

	if !param.tlsEnabled {
		log.Debug("TLS disabled")
		return grpc.NewClient(addr,
			grpc.WithUnaryInterceptor(
				cookie.CookieInterceptor(cookie.NewJar(), addr)),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	log.Debug("TLS enabled")
	// ok, we need to use TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	if param.cert != "" && param.key != "" {
		cert, err := tls.LoadX509KeyPair(param.cert, param.key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if param.ca != "" {
		caCert, err := os.ReadFile(param.ca)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
			return nil, fmt.Errorf("failed to append server certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}
	if param.skipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	return grpc.NewClient(addr,
		grpc.WithUnaryInterceptor(
			cookie.CookieInterceptor(cookie.NewJar(), addr)),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
}

func ConnectGrpc(cfg *config.CliArgs) (*grpc.ClientConn, error) {
	return NewClient(cfg.Addr, WithCliArgs(cfg))
}
