package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/ShareChat/moj-feed-go-lib/option"
	tfs "github.com/ShareChat/tardis-feature-service-protocol/models/v1/featureservice"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpcpool "github.com/processout/grpc-go-pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	connectionIdleTimeout     = 2 * time.Minute
	connectionMaxLifeDuration = 3 * time.Minute
)

type FeatureSetParams struct {
	FeatureSetType    string
	FeatureSetId      string
	FeatureSetVersion int32
	FeatureClass      tfs.TardisGetFeaturesRequest_FeatureClass
}

type CounterFeatures map[string]option.Option[any]

func createGrpcTardisFeatureService(servingAddress string) *grpcpool.Pool {
	//var factory grpcpool.Factory
	factory := func() (*grpc.ClientConn, error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
		defer cancel()

		var tlsConf tls.Config
		tlsConf.InsecureSkipVerify = true
		creds := credentials.NewTLS(&tlsConf)

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
			grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor),
			grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024)),
		}

		grpc_prometheus.EnableClientHandlingTimeHistogram()

		conn, err := grpc.DialContext(ctx, servingAddress, opts...)
		if err != nil {
			log.Printf("Failed to start gRPC connection, err=%+v", err)
		}
		return conn, err
	}

	// Generally, one pod handles 100 req/sec, 100 connections in a pool should be good enough
	pool, err := grpcpool.New(factory, 100, 100, connectionIdleTimeout, connectionMaxLifeDuration)
	if err != nil {
		log.Fatalf("failed to create gRPC pool: %v", err)
	}

	return pool
}

// Function to use the connection pool to make a gRPC call to the Tardis Feature service
func callTardisFeatureService(ctx context.Context, pool *grpcpool.Pool, request *tfs.TardisGetFeaturesRequest) (*tfs.TardisGetFeaturesResponse, error) {

	conn, err := pool.Get(ctx)
	defer func(conn *grpcpool.ClientConn) {
		errClose := conn.Close()
		if errClose != nil {
			fmt.Print("Error closing grpc connection", "errClose", errClose)
		}
	}(conn)
	if err != nil {
		return nil, err
	}

	client := tfs.NewTardisFeatureServiceClient(conn)

	response, err := client.GetFeatures(ctx, request /*grpc.CallContentSubtype(tfs_grpc_codec.GrpcCodecName)*/)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	pool := createGrpcTardisFeatureService("live-tardis-feature-service.sharechat.internal:443")

	chatroomIds := readChatroomIds()

	parsedResponseInOneGo := fetchFeaturesInOneGo(ctx, pool, chatroomIds)
	parsedResponseParallel := fetchFeaturesParallel(ctx, pool, chatroomIds)

	saveResponseToFIle(parsedResponseInOneGo, "oneGo.json")
	saveResponseToFIle(parsedResponseParallel, "parallel.json")
	//file1, _ := os.Create("oneGo.txt")
	//file2, _ := os.Create("parallel.txt")
	//fmt.Fprintln(file1, parsedResponseInOneGo)
	//fmt.Fprintln(file2, parsedResponseParallel)
}
