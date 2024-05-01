package metrics

import (
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"time"
)

type MetricProvider interface {
	GetMetrics(
		region string,
		namespace string,
		metricNames []string,
		filters map[string][]string,
		startTime, endTime time.Time,
		interval time.Duration,
		statistics []types2.Statistic,
	) (map[string][]types2.Datapoint, error)
}
