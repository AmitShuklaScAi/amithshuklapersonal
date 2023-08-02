package main

import tfs "github.com/ShareChat/tardis-feature-service-protocol/models/v1/featureservice"

var CrossFeatures = []*tfs.TardisFeatureRequest{
	{
		FeatureName: "giftAmount",
		FeatureType: tfs.TardisFeatureRequest_DOUBLE,
		AggregationRanges: []*tfs.AggregationRange{
			{
				Count: 15,
				Unit:  tfs.AggregationRange_MINUTE,
			},
			{
				Count: 1,
				Unit:  tfs.AggregationRange_HOUR,
			},
			{
				Count: 6,
				Unit:  tfs.AggregationRange_HOUR,
			},

			{
				Count: 1,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 3,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 7,
				Unit:  tfs.AggregationRange_DAY,
			},
		},
	},
	{
		FeatureName: "commentCount",
		FeatureType: tfs.TardisFeatureRequest_DOUBLE,
		AggregationRanges: []*tfs.AggregationRange{
			{
				Count: 15,
				Unit:  tfs.AggregationRange_MINUTE,
			},
			{
				Count: 1,
				Unit:  tfs.AggregationRange_HOUR,
			},
			{
				Count: 6,
				Unit:  tfs.AggregationRange_HOUR,
			},

			{
				Count: 1,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 3,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 7,
				Unit:  tfs.AggregationRange_DAY,
			},
		},
	},
	{
		FeatureName: "requestAudioSeatCount",
		FeatureType: tfs.TardisFeatureRequest_DOUBLE,
		AggregationRanges: []*tfs.AggregationRange{
			{
				Count: 15,
				Unit:  tfs.AggregationRange_MINUTE,
			},
			{
				Count: 1,
				Unit:  tfs.AggregationRange_HOUR,
			},
			{
				Count: 6,
				Unit:  tfs.AggregationRange_HOUR,
			},

			{
				Count: 1,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 3,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 7,
				Unit:  tfs.AggregationRange_DAY,
			},
		},
	},
	{
		FeatureName: "GmvAmount",
		FeatureType: tfs.TardisFeatureRequest_DOUBLE,
		AggregationRanges: []*tfs.AggregationRange{
			{
				Count: 15,
				Unit:  tfs.AggregationRange_MINUTE,
			},
			{
				Count: 1,
				Unit:  tfs.AggregationRange_HOUR,
			},
			{
				Count: 6,
				Unit:  tfs.AggregationRange_HOUR,
			},

			{
				Count: 1,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 3,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 7,
				Unit:  tfs.AggregationRange_DAY,
			},
		},
	},
	{
		FeatureName: "chatroomJoined",
		FeatureType: tfs.TardisFeatureRequest_DOUBLE,
		AggregationRanges: []*tfs.AggregationRange{
			{
				Count: 15,
				Unit:  tfs.AggregationRange_MINUTE,
			},
			{
				Count: 1,
				Unit:  tfs.AggregationRange_HOUR,
			},
			{
				Count: 6,
				Unit:  tfs.AggregationRange_HOUR,
			},

			{
				Count: 1,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 3,
				Unit:  tfs.AggregationRange_DAY,
			},
			{
				Count: 7,
				Unit:  tfs.AggregationRange_DAY,
			},
		},
	},
}
