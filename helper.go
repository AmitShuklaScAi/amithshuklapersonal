package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/ShareChat/moj-feed-go-lib/option"
	tfs "github.com/ShareChat/tardis-feature-service-protocol/models/v1/featureservice"
	"golang.org/x/sync/errgroup"
)

const (
	compositeEntityDelimiter = "compositeEntityDelimiter"
)

func getAggregationKey(aggregationRange *tfs.AggregationRange) (string, error) {
	var aggregationKey string
	if aggregationRange.GetUnit() == tfs.AggregationRange_LIFETIME {
		aggregationKey = aggregationRange.GetUnit().String()
	} else {
		aggregationKey = strconv.Itoa(int(aggregationRange.GetCount())) + "_" + aggregationRange.GetUnit().String()
	}

	return aggregationKey, nil
	//if aggregationKey == Agg1hr || aggregationKey == Agg1day || aggregationKey == Agg7day || aggregationKey == Agg30day {
	//	return aggregationKey, nil
	//}
	//return "", fmt.Errorf("invalid request. Unknown aggregation range %v", aggregationKey)
}

func getFeatureValue(value interface{}) (interface{}, error) {
	switch val := value.(type) {
	case float64:
		intFloor := int64(math.Floor(val))
		intCeil := int64(math.Ceil(val))
		if intCeil != intFloor {
			return 0, fmt.Errorf("expected value of type integer got - %v, of type float", value)
		}
		return intFloor, nil
	case string:
		return val, nil
	case int64:
		return val, nil
	case int32:
		return val, nil
	default:
		return nil, nil
	}
}

func setFeatures(counterFeatures CounterFeatures, featureKey string, value option.Option[any]) {
	counterFeatures[featureKey] = value
}

func transformPointerToValueMap(counterFeaturesMap map[string]CounterFeatures) map[string]CounterFeatures {
	valueMap := make(map[string]CounterFeatures, len(counterFeaturesMap))
	for k, v := range counterFeaturesMap {
		valueMap[k] = v
	}
	return valueMap
}

func chunkListByEntityId(responses []*tfs.TardisFeatureResponse, numChunks int) [][]*tfs.TardisFeatureResponse {
	result := make([][]*tfs.TardisFeatureResponse, numChunks)
	entityIdToChunkIdxMap := make(map[string]int, 0)

	for _, response := range responses {
		entityId := strings.Join([]string{response.Filter.L2Filter}, compositeEntityDelimiter)
		var chunkIdx int
		var ok bool
		if chunkIdx, ok = entityIdToChunkIdxMap[entityId]; !ok {
			chunkIdx = rand.Intn(numChunks)
			entityIdToChunkIdxMap[entityId] = chunkIdx
		}
		result[chunkIdx] = append(result[chunkIdx], response)
	}

	return result
}

func parseTardisFeatureResponse(ctx context.Context, tardisResponse *tfs.TardisGetFeaturesResponse) (map[string]CounterFeatures, error) {

	counterFeaturesMap := make(map[string]CounterFeatures, 0)
	parseResponseConcurrency := 5
	chunkedFeatureResponses := chunkListByEntityId(tardisResponse.FeatureResponse, parseResponseConcurrency)
	responseChan := make(chan map[string]CounterFeatures, parseResponseConcurrency)

	eg := new(errgroup.Group)

	for _, featureResponseChunk := range chunkedFeatureResponses {
		featureResponseChunk := featureResponseChunk // https://golang.org/doc/faq#closures_and_goroutines
		eg.Go(func() error {
			err := parseTardisFeatureResponseForChunk(featureResponseChunk, responseChan)
			return err
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	close(responseChan)

	for resp := range responseChan {
		partialResult := resp
		for id, feature := range partialResult {
			counterFeaturesMap[id] = feature
		}
	}

	return counterFeaturesMap, nil
}

func parseTardisFeatureResponseForChunk(featureResponseChunk []*tfs.TardisFeatureResponse,
	responseChannel chan<- map[string]CounterFeatures) error {
	counterFeaturesMap := make(map[string]CounterFeatures, 0)
	var entityId string
	for _, featureResponse := range featureResponseChunk {

		entityId = strings.Join([]string{featureResponse.Filter.L2Filter}, "compositeEntityDelimiter")

		featureName := featureResponse.FeatureName
		var exists bool
		var counterFeatures CounterFeatures
		if counterFeatures, exists = counterFeaturesMap[entityId]; !exists {
			counterFeatures = CounterFeatures{}
			counterFeaturesMap[entityId] = counterFeatures
		}

		var featureValue option.Option[any]
		var featureValueErr error
		var value interface{}
		tardisErr := featureResponse.FeatureValue.GetError()
		if tardisErr != nil {
			if tardisErr.ErrorType == tfs.TardisResponseError_NOT_FOUND {
				// Zero-counters are not stored in Tardis and therefore result in a NOT_FOUND error
				// This is a special case which we consider to represent zero-counters
				featureValue = option.Some[any](0)
			} else {
				// Expected Tardis Error - skip this response
				// These will be represented as None (ie null in logs)
				continue
			}
		} else {
			switch featureResponse.FeatureValue.GetValue().(type) {
			case *tfs.FeatureValue_DoubleValue:
				tardisValue := featureResponse.FeatureValue.GetValue().(*tfs.FeatureValue_DoubleValue)
				value, featureValueErr = getFeatureValue(tardisValue.DoubleValue)
			case *tfs.FeatureValue_IntValue:
				tardisValue := featureResponse.FeatureValue.GetValue().(*tfs.FeatureValue_IntValue)
				value, featureValueErr = getFeatureValue(tardisValue.IntValue)
			case *tfs.FeatureValue_StringValue:
				tardisValue := featureResponse.FeatureValue.GetValue().(*tfs.FeatureValue_StringValue)
				value, featureValueErr = getFeatureValue(tardisValue.StringValue)
			case *tfs.FeatureValue_StringList:
				tardisValue := featureResponse.FeatureValue.GetValue().(*tfs.FeatureValue_StringList)
				value, featureValueErr = getFeatureValue(tardisValue.StringList)
			}

			if featureValueErr != nil {
				return featureValueErr
			}
			featureValue = option.Some(value)
		}

		aggregationKey, aggErr := getAggregationKey(featureResponse.AggregationRange)
		if aggErr != nil {
			// Fatal - invalid aggregation range
			return aggErr
		}

		var featureKey string
		var prefix string = "cross"
		featureKey = fmt.Sprintf("%s_%s_%s", prefix, featureName, aggregationKey)

		setFeatures(counterFeatures, featureKey, featureValue)
	}
	responseChannel <- transformPointerToValueMap(counterFeaturesMap)
	return nil
}

func buildRequestCrossFeatures(chatroomIds []string) *tfs.TardisGetFeaturesRequest {

	userId := []string{"980537515"}

	requestParams := FeatureSetParams{
		FeatureSetType:    "user-chatroom",
		FeatureSetId:      "user_chatroom_cross_counter_features",
		FeatureSetVersion: 1,
		FeatureClass:      1,
	}

	tardisEntityVariants := make([]*tfs.TardisEntityVariant, len(userId))

	for idx, entityId := range userId {
		tardisEntityVariants[idx] = &tfs.TardisEntityVariant{
			CompositeEntityId: &tfs.TardisCompositeEntityId{EntityIdComponents: []string{entityId}},
		}
	}

	filters := buildFiltersForCrossFeatures(chatroomIds)

	for idx := range CrossFeatures {
		CrossFeatures[idx].Filters = filters
	}

	// Last, we put these together in a request struct with feature set details
	req := &tfs.TardisGetFeaturesRequest{
		EntityVariants:    tardisEntityVariants,
		FeatureSetId:      requestParams.FeatureSetId,
		FeatureSetVersion: requestParams.FeatureSetVersion,
		FeatureSetType:    requestParams.FeatureSetType,
		FeatureClass:      requestParams.FeatureClass,
		Features:          CrossFeatures,
	}

	return req

}

func readChatroomIds() []string {
	file, err := os.Open("cr.txt")
	if err != nil {
		fmt.Println("Error opening the file:", err)
		return nil
	}
	defer file.Close()

	var chatroomIds []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		chatroomIds = append(chatroomIds, scanner.Text())
	}

	return chatroomIds
}

func buildFiltersForCrossFeatures(chatroomIds []string) []*tfs.Filter {

	filters := make([]*tfs.Filter, len(chatroomIds))

	for idx, cId := range chatroomIds {
		filters[idx] = &tfs.Filter{
			L1Filter: "crossFeatures",
			L2Filter: cId,
		}
	}

	return filters
}

func saveResponseToFIle(response map[string]CounterFeatures, fname string) {

	jsonData, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return
	}

	if !json.Valid(jsonData) {
		fmt.Println("Invalid JSON data")
		return
	}

	// Save the JSON data to a file.
	file, err := os.Create(fname)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing JSON data to file:", err)
		return
	}
}
