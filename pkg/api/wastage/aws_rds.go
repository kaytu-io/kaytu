package wastage

import types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"

type AwsRdsClusterType string

const (
	AwsRdsClusterTypeSingleInstance     AwsRdsClusterType = "Single-AZ"
	AwsRdsClusterTypeMultiAzOneInstance AwsRdsClusterType = "Multi-AZ"
	AwsRdsClusterTypeMultiAzTwoInstance AwsRdsClusterType = "Multi-AZ (readable standbys)"
)

type AwsRds struct {
	HashedInstanceId                   string            `json:"hashedInstanceId"`
	AvailabilityZone                   string            `json:"availabilityZone"`
	InstanceType                       string            `json:"instanceType"`
	Engine                             string            `json:"engine"`
	EngineVersion                      string            `json:"engineVersion"`
	LicenseModel                       string            `json:"licenseModel"`
	BackupRetentionPeriod              *int32            `json:"backupRetentionPeriod"`
	ClusterType                        AwsRdsClusterType `json:"clusterType"`
	PerformanceInsightsEnabled         bool              `json:"performanceInsightsEnabled"`
	PerformanceInsightsRetentionPeriod *int32            `json:"performanceInsightsRetentionPeriod"`

	StorageType       *string `json:"storageType"`
	StorageSize       *int32  `json:"storageSize"`
	StorageIops       *int32  `json:"storageIops"`
	StorageThroughput *int32  `json:"storageThroughput"`
}

type RightsizingAwsRds struct {
	Region        string            `json:"region"`
	InstanceType  string            `json:"instanceType"`
	Engine        string            `json:"engine"`
	EngineVersion string            `json:"engineVersion"`
	ClusterType   AwsRdsClusterType `json:"clusterType"`

	VCPU     int64 `json:"vCPU"`
	MemoryGb int64 `json:"memoryGb"`

	StorageType       *string `json:"storageType"`
	StorageSize       *int32  `json:"storageSize"`
	StorageIops       *int32  `json:"storageIops"`
	StorageThroughput *int32  `json:"storageThroughput"`

	Cost float64 `json:"cost"`
}

type AwsRdsRightsizingRecommendation struct {
	Current     RightsizingAwsRds  `json:"current"`
	Recommended *RightsizingAwsRds `json:"recommended"`

	Description string `json:"description"`
}

type AwsRdsWastageRequest struct {
	Instance    AwsRds                        `json:"instance"`
	Metrics     map[string][]types2.Datapoint `json:"metrics"`
	Region      string                        `json:"region"`
	Preferences map[string]*string            `json:"preferences"`
}

type AwsRdsWastageResponse struct {
	RightSizing AwsRdsRightsizingRecommendation `json:"rightSizing"`
}
