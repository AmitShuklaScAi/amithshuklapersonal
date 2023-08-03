package main

import (
	"context"
	"fmt"
	"sync"

	grpcpool "github.com/processout/grpc-go-pool"
	"golang.org/x/sync/errgroup"
)

func fetchFeaturesParallel(ctx context.Context, pool *grpcpool.Pool, chatRoomIDs []string) map[string]CounterFeatures {

	g, _ := errgroup.WithContext(ctx)

	batchSize := 100
	batchedChatroomIdsLen := len(chatRoomIDs) / batchSize
	if !(len(chatRoomIDs)%batchSize == 0) {
		batchedChatroomIdsLen++
	}
	groupBatchedChatroomIDs := make([][]string, batchedChatroomIdsLen)
	for idx, chatRoomID := range chatRoomIDs {
		batchNum := idx / batchSize
		groupBatchedChatroomIDs[batchNum] = append(groupBatchedChatroomIDs[batchNum], chatRoomID)
	}

	crossFeatureMap := make(map[string]CounterFeatures, len(chatRoomIDs))
	crossFeatureMapLock := &sync.Mutex{}

	for _, crossbatchedChatroomIDs := range groupBatchedChatroomIDs {

		batchedChatroomIdsCpy := crossbatchedChatroomIDs

		g.Go(func() error {

			request := buildRequestCrossFeatures(batchedChatroomIdsCpy)

			response, err := callTardisFeatureService(ctx, pool, request)
			if err != nil {
				fmt.Printf("Error calling Tardis Feature service in parallel: %v", err)
				return err
			}
			crossFeatures, err := parseTardisFeatureResponse(ctx, response)
			if err != nil {
				fmt.Printf("Error parsing TFS respponse: %v", err)
				return err
			}

			crossFeatureMapLock.Lock()
			for crosschatroomID, crossFeature := range crossFeatures {
				crossFeatureMap[crosschatroomID] = crossFeature
			}

			crossFeatureMapLock.Unlock()
			return nil
		})
		//time.Sleep(time.Millisecond * 1)
	}

	err := g.Wait()

	if err != nil {
		fmt.Println("Error in fetching TFS features in parallel: ", err)
	}
	return crossFeatureMap
}

func fetchFeaturesInOneGo(ctx context.Context, pool *grpcpool.Pool, chatroomIds []string) map[string]CounterFeatures {

	request := buildRequestCrossFeatures(chatroomIds)

	response, err := callTardisFeatureService(ctx, pool, request)
	if err != nil {
		fmt.Printf("Error calling Tardis Feature service in one go: %v", err)
	}

	parsedResponse, err := parseTardisFeatureResponse(ctx, response)
	if err != nil {
		fmt.Printf("Error parsing TFS respponse: %v", err)
	}

	return parsedResponse
}
