package metrics

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"time"
)

type CloudWatch struct {
	cfg aws.Config
}

func NewCloudWatch(cfg aws.Config) (*CloudWatch, error) {
	return &CloudWatch{cfg: cfg}, nil
}

func (cw *CloudWatch) GetMetrics(
	region string,
	namespace string,
	metricNames []string,
	filters map[string][]string,
	startTime, endTime time.Time,
	interval time.Duration,
	statistics []types2.Statistic,
) (map[string][]types2.Datapoint, error) {
	localCfg := cw.cfg
	localCfg.Region = region

	metrics := map[string][]types2.Datapoint{}

	ctx := context.Background()
	cloudwatchClient := cloudwatch.NewFromConfig(localCfg)
	var dimensionFilters []types2.DimensionFilter
	for k, v := range filters {
		dimensionFilters = append(dimensionFilters, types2.DimensionFilter{
			Name:  aws.String(k),
			Value: aws.String(v[0]),
		})
	}

	var dimensions []types2.Dimension
	for k, v := range filters {
		dimensions = append(dimensions, types2.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v[0]),
		})
	}

	paginator := cloudwatch.NewListMetricsPaginator(cloudwatchClient, &cloudwatch.ListMetricsInput{
		Namespace:  aws.String(namespace),
		Dimensions: dimensionFilters,
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, p := range page.Metrics {
			if p.MetricName == nil {
				continue
			}

			exists := false
			for _, mn := range metricNames {
				if *p.MetricName == mn {
					exists = true
					break
				}
			}

			if !exists {
				continue
			}

			// Create input for GetMetricStatistics
			input := &cloudwatch.GetMetricStatisticsInput{
				Namespace:  aws.String(namespace),
				MetricName: p.MetricName,
				Dimensions: dimensions,
				StartTime:  aws.Time(startTime),
				EndTime:    aws.Time(endTime),
				Period:     aws.Int32(int32(interval.Seconds())),
				Statistics: statistics,
			}

			// Get metric data
			resp, err := cloudwatchClient.GetMetricStatistics(ctx, input)
			if err != nil {
				return nil, err
			}

			metrics[*p.MetricName] = resp.Datapoints
		}
	}
	return metrics, nil
}
