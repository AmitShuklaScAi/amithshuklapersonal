package main

import (
	"crypto/tls"

	tfs "github.com/ShareChat/tardis-feature-service-protocol/models/v1/featureservice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TFSClient(URL string) tfs.TardisFeatureServiceClient {
	var conn *grpc.ClientConn
	var tlsConf tls.Config
	tlsConf.InsecureSkipVerify = true
	// TODO:  Check why do we need to remove auth
	creds := credentials.NewTLS(&tlsConf)
	conn, err := grpc.Dial(URL, grpc.WithTransportCredentials(creds))
	if err != nil {
		panic(err)
	}
	c := tfs.NewTardisFeatureServiceClient(conn)
	return c
}
