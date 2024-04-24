package wastage

import (
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2Placement struct {
	Tenancy          types.Tenancy `json:"tenancy"`
	AvailabilityZone string        `json:"availabilityZone"`
	HashedHostId     string        `json:"hashedHostId"`
}

type EC2Instance struct {
	HashedInstanceId  string                      `json:"hashedInstanceId"`
	State             types.InstanceStateName     `json:"state"`
	InstanceType      types.InstanceType          `json:"instanceType"`
	Platform          string                      `json:"platform"`
	ThreadsPerCore    int32                       `json:"threadsPerCore"`
	CoreCount         int32                       `json:"coreCount"`
	EbsOptimized      bool                        `json:"ebsOptimized"`
	InstanceLifecycle types.InstanceLifecycleType `json:"instanceLifecycle"`
	Monitoring        *types.MonitoringState      `json:"monitoring"`
	Placement         *EC2Placement               `json:"placement"`
}

type EC2Volume struct {
	HashedVolumeId   string           `json:"hashedVolumeId"`
	VolumeType       types.VolumeType `json:"volumeType"`
	Size             *int32           `json:"size"`
	Iops             *int32           `json:"iops"`
	AvailabilityZone *string          `json:"availabilityZone"`
	Throughput       *float64         `json:"throughput"`
}

type EC2InstanceWastageRequest struct {
	HashedAccountID string                                   `json:"hashedAccountID"`
	HashedUserID    string                                   `json:"hashedUserID"`
	HashedARN       string                                   `json:"hashedARN"`
	Instance        EC2Instance                              `json:"instance"`
	Volumes         []EC2Volume                              `json:"volumes"`
	Metrics         map[string][]types2.Datapoint            `json:"metrics"`
	VolumeMetrics   map[string]map[string][]types2.Datapoint `json:"volumeMetrics"`
	Region          string                                   `json:"region"`
	Preferences     map[string]*string                       `json:"preferences"`
}

type RightSizingRecommendation struct {
	TargetInstanceType string  `json:"targetInstanceType"`
	Saving             float64 `json:"saving"`
	CurrentCost        float64 `json:"currentCost"`
	TargetCost         float64 `json:"targetCost"`

	AvgCPUUsage string `json:"avgCPUUsage"`
	TargetCores string `json:"targetCores"`

	AvgNetworkBandwidth       string `json:"avgNetworkBandwidth"`
	TargetNetworkPerformance  string `json:"targetNetworkBandwidth"`
	CurrentNetworkPerformance string `json:"currentNetworkPerformance"`

	TargetEBSBandwidth  string `json:"targetEBSBandwidth"`
	CurrentEBSBandwidth string `json:"currentEBSBandwidth"`
	AvgEBSBandwidth     string `json:"avgEBSBandwidth"`

	MaxMemoryUsagePercentage string `json:"maxMemoryUsagePercentage"`
	CurrentMemory            string `json:"currentMemory"`
	TargetMemory             string `json:"targetMemory"`

	VolumesCurrentSizes map[string]int32 `json:"volumeCurrentSizes"`
	VolumesTargetSizes  map[string]int32 `json:"volumeTargetSizes"`

	VolumesCurrentTypes map[string]types.VolumeType `json:"volumeCurrentTypes"`
	VolumesTargetTypes  map[string]types.VolumeType `json:"volumeTargetTypes"`

	VolumesCurrentIOPS        map[string]int32   `json:"volumeCurrentIOPS"`
	VolumesTargetBaselineIOPS map[string]int32   `json:"volumeTargetBaselineIOPS"`
	VolumesTargetIOPS         map[string]int32   `json:"volumeTargetIOPS"`
	AvgVolumesIOPSUtilization map[string]float64 `json:"avgVolumesIOPSUtilization"`
	MinVolumesIOPSUtilization map[string]float64 `json:"minVolumesIOPSUtilization"`
	MaxVolumesIOPSUtilization map[string]float64 `json:"maxVolumesIOPSUtilization"`

	VolumesCurrentThroughput        map[string]float64 `json:"volumeCurrentThroughput"`
	VolumesTargetBaselineThroughput map[string]float64 `json:"volumeTargetBaselineThroughput"`
	VolumesTargetThroughput         map[string]float64 `json:"volumeTargetThroughput"`
	AvgVolumesThroughputUtilization map[string]float64 `json:"avgVolumesThroughputUtilization"`
	MinVolumesThroughputUtilization map[string]float64 `json:"minVolumesThroughputUtilization"`
	MaxVolumesThroughputUtilization map[string]float64 `json:"maxVolumesThroughputUtilization"`

	VolumesCurrentCosts map[string]float64 `json:"volumeCurrentCosts"`
	VolumesTargetCosts  map[string]float64 `json:"volumeTargetCosts"`
	VolumesSaving       map[string]float64 `json:"volumeSaving"`

	Description string `json:"description"`
}

type EC2InstanceWastageResponse struct {
	CurrentCost     float64                    `json:"currentCost"`
	TotalSavings    float64                    `json:"totalSavings"`
	EbsTotalSavings map[string]float64         `json:"ebsTotalSavings"`
	RightSizing     *RightSizingRecommendation `json:"rightSizing"`
}
