package processor

import (
	"fmt"
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	preferences2 "github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/cmd/optimize/style"
	"github.com/kaytu-io/kaytu/cmd/optimize/view"
	"github.com/kaytu-io/kaytu/pkg/api/wastage"
)

type RDSInstanceItem struct {
	Instance            types.DBInstance
	Region              string
	OptimizationLoading bool
	Preferences         []preferences2.PreferenceItem
	Skipped             bool
	SkipReason          *string

	Metrics map[string][]types2.Datapoint
	Wastage wastage.AwsRdsWastageResponse
}

func (i RDSInstanceItem) RDSInstanceDevice() view.Device {
	ec2Instance := view.Device{
		Properties:   nil,
		DeviceID:     *i.Instance.DBInstanceIdentifier,
		ResourceType: "RDS Instance",
		Runtime:      "730 hours",
		CurrentCost:  i.Wastage.RightSizing.Current.Cost,
	}
	regionProperty := view.Property{
		Key:     style.Bold.Render("Region"),
		Current: i.Wastage.RightSizing.Current.Region,
	}
	instanceSizeProperty := view.Property{
		Key:     style.Bold.Render("Instance Size"),
		Current: i.Wastage.RightSizing.Current.InstanceType,
	}
	engineProperty := view.Property{
		Key:     style.Bold.Render("Engine"),
		Current: i.Wastage.RightSizing.Current.Engine,
	}
	engineVerProperty := view.Property{
		Key:     style.Bold.Render("Engine Version"),
		Current: i.Wastage.RightSizing.Current.EngineVersion,
	}
	clusterTypeProperty := view.Property{
		Key:     style.Bold.Render("Cluster Type"),
		Current: string(i.Wastage.RightSizing.Current.ClusterType),
	}
	vCPUProperty := view.Property{
		Key:     "  vCPU",
		Current: fmt.Sprintf("%d", i.Wastage.RightSizing.Current.VCPU),
		Average: view.Percentage(i.Wastage.RightSizing.VCPU.Avg),
		Max:     view.Percentage(i.Wastage.RightSizing.VCPU.Max),
	}
	memoryProperty := view.Property{
		Key:     "  Memory",
		Current: fmt.Sprintf("%d GiB", i.Wastage.RightSizing.Current.MemoryGb),
		Average: view.MemoryUsagePercentageByFreeSpace(i.Wastage.RightSizing.FreeMemoryBytes.Avg, float64(i.Wastage.RightSizing.Current.MemoryGb)),
		Max:     view.MemoryUsagePercentageByFreeSpace(i.Wastage.RightSizing.FreeMemoryBytes.Min, float64(i.Wastage.RightSizing.Current.MemoryGb)),
	}
	storageTypeProperty := view.Property{
		Key:     "  Type",
		Current: view.PString(i.Wastage.RightSizing.Current.StorageType),
	}
	storageSizeProperty := view.Property{
		Key:     "  Size",
		Current: view.SizeByteToGB(i.Wastage.RightSizing.Current.StorageSize),
		Average: view.StorageUsagePercentageByFreeSpace(i.Wastage.RightSizing.FreeStorageBytes.Avg, i.Wastage.RightSizing.Current.StorageSize),
		Max:     view.StorageUsagePercentageByFreeSpace(i.Wastage.RightSizing.FreeStorageBytes.Min, i.Wastage.RightSizing.Current.StorageSize),
	}
	storageIOPSProperty := view.Property{
		Key:     "  IOPS",
		Current: fmt.Sprintf("%d", i.Wastage.RightSizing.Current.StorageIops),
		Average: fmt.Sprintf("%s io/s", view.PFloat64ToString(i.Wastage.RightSizing.StorageIops.Avg)),
		Max:     fmt.Sprintf("%s io/s", view.PFloat64ToString(i.Wastage.RightSizing.StorageIops.Max)),
	}
	storageThroughputProperty := view.Property{
		Key:     "  Throughput",
		Current: view.PStorageThroughputMbps(i.Wastage.RightSizing.Current.StorageThroughput),
		Average: view.PStorageThroughputMbps(i.Wastage.RightSizing.StorageThroughputBytes.Avg),
		Max:     view.PStorageThroughputMbps(i.Wastage.RightSizing.StorageThroughputBytes.Max),
	}

	if i.Wastage.RightSizing.Recommended != nil {
		ec2Instance.RightSizedCost = i.Wastage.RightSizing.Recommended.Cost
		regionProperty.Recommended = i.Wastage.RightSizing.Recommended.Region
		instanceSizeProperty.Recommended = i.Wastage.RightSizing.Recommended.InstanceType
		engineProperty.Recommended = i.Wastage.RightSizing.Recommended.Engine
		engineVerProperty.Recommended = i.Wastage.RightSizing.Recommended.EngineVersion
		clusterTypeProperty.Recommended = string(i.Wastage.RightSizing.Recommended.ClusterType)
		vCPUProperty.Recommended = fmt.Sprintf("%d", i.Wastage.RightSizing.Recommended.VCPU)
		memoryProperty.Recommended = fmt.Sprintf("%d GiB", i.Wastage.RightSizing.Recommended.MemoryGb)
		storageTypeProperty.Recommended = view.PString(i.Wastage.RightSizing.Recommended.StorageType)
		storageSizeProperty.Recommended = view.SizeByteToGB(i.Wastage.RightSizing.Recommended.StorageSize)
		storageIOPSProperty.Recommended = fmt.Sprintf("%s io/s", view.PInt32ToString(i.Wastage.RightSizing.Recommended.StorageIops))
		storageThroughputProperty.Recommended = view.PStorageThroughputMbps(i.Wastage.RightSizing.Recommended.StorageThroughput)
	}
	ec2Instance.Properties = append(ec2Instance.Properties, regionProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, instanceSizeProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, engineProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, engineVerProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, clusterTypeProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, view.Property{
		Key: style.Bold.Render("Compute"),
	})
	ec2Instance.Properties = append(ec2Instance.Properties, vCPUProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, memoryProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, view.Property{
		Key: style.Bold.Render("Storage"),
	})
	ec2Instance.Properties = append(ec2Instance.Properties, storageTypeProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, storageSizeProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, storageIOPSProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, storageThroughputProperty)

	return ec2Instance
}

func (i RDSInstanceItem) Devices() []view.Device {
	return []view.Device{i.RDSInstanceDevice()}
}

func (i RDSInstanceItem) ToOptimizationItem() view.OptimizationItem {
	oi := view.OptimizationItem{
		ID:           *i.Instance.DBInstanceIdentifier,
		ResourceType: *i.Instance.DBInstanceClass,
		Region:       i.Region,
		Devices:      i.Devices(),
		Preferences:  i.Preferences,
		Description:  i.Wastage.RightSizing.Description,
		Loading:      i.OptimizationLoading,
		Skipped:      i.Skipped,
		SkipReason:   i.SkipReason,
	}

	//if i.Instance.PlatformDetails != nil {
	//	oi.Platform = *i.Instance.PlatformDetails
	//}
	//for _, t := range i.Instance.Tags {
	//	if t.Key != nil && strings.ToLower(*t.Key) == "name" && t.Value != nil {
	//		oi.Name = *t.Value
	//	}
	//}
	if oi.Name == "" {
		oi.Name = *i.Instance.DBInstanceIdentifier
	}

	return oi
}
