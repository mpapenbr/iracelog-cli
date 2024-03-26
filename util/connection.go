package util

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectGRPC(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	return conn, err
}
