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
	vCPUProperty := view.Property{
		Key:     "  vCPU",
		Current: fmt.Sprintf("%d", i.Wastage.RightSizing.Current.VCPU),
		//Average: view.Percentage(i.Wastage.RightSizing.VCPU.Avg),
		//Max:     view.Percentage(i.Wastage.RightSizing.VCPU.Max),
	}
	processorProperty := view.Property{
		Key: "  Processor(s)",
		//Current: i.Wastage.RightSizing.Current.Processor,
	}
	architectureProperty := view.Property{
		Key: "  Architecture",
		//Current: i.Wastage.RightSizing.Current.Architecture,
	}
	licenseCostProperty := view.Property{
		Key: "  License Cost",
		//Current: fmt.Sprintf("$%.2f", i.Wastage.RightSizing.Current.LicensePrice),
	}
	memoryProperty := view.Property{
		Key:     "  Memory",
		Current: fmt.Sprintf("%d GiB", i.Wastage.RightSizing.Current.MemoryGb),
		//Average: view.Percentage(i.Wastage.RightSizing.Memory.Avg),
		//Max:     view.Percentage(i.Wastage.RightSizing.Memory.Max),
	}
	ebsProperty := view.Property{
		Key: "EBS Bandwidth",
		//Current: fmt.Sprintf("%s", i.Wastage.RightSizing.Current.EBSBandwidth),
		//Average: view.PNetworkThroughputMbps(i.Wastage.RightSizing.EBSBandwidth.Avg),
		//Max:     view.PNetworkThroughputMbps(i.Wastage.RightSizing.EBSBandwidth.Max),
	}
	iopsProperty := view.Property{
		Key: "EBS IOPS",
		//Current: fmt.Sprintf("%s", i.Wastage.RightSizing.Current.EBSIops),
		//Average: fmt.Sprintf("%s io/s", view.PFloat64ToString(i.Wastage.RightSizing.EBSIops.Avg)),
		//Max:     fmt.Sprintf("%s io/s", view.PFloat64ToString(i.Wastage.RightSizing.EBSIops.Max)),
	}
	netThroughputProperty := view.Property{
		Key: "  Throughput",
		//Current: fmt.Sprintf("%s", i.Wastage.RightSizing.Current.NetworkThroughput),
		//Average: view.PNetworkThroughputMbps(i.Wastage.RightSizing.NetworkThroughput.Avg),
		//Max:     view.PNetworkThroughputMbps(i.Wastage.RightSizing.NetworkThroughput.Max),
	}
	enaProperty := view.Property{
		Key: "  ENA",
		//Current: fmt.Sprintf("%s", i.Wastage.RightSizing.Current.ENASupported),
	}

	if i.Wastage.RightSizing.Recommended != nil {
		ec2Instance.RightSizedCost = i.Wastage.RightSizing.Recommended.Cost
		regionProperty.Recommended = i.Wastage.RightSizing.Recommended.Region
		instanceSizeProperty.Recommended = i.Wastage.RightSizing.Recommended.InstanceType
		vCPUProperty.Recommended = fmt.Sprintf("%d", i.Wastage.RightSizing.Recommended.VCPU)
		//processorProperty.Recommended = i.Wastage.RightSizing.Recommended.Processor
		//architectureProperty.Recommended = i.Wastage.RightSizing.Recommended.Architecture
		//licenseCostProperty.Recommended = fmt.Sprintf("$%.2f", i.Wastage.RightSizing.Recommended.LicensePrice)
		memoryProperty.Recommended = fmt.Sprintf("%d GiB", i.Wastage.RightSizing.Recommended.MemoryGb)
		//ebsProperty.Recommended = i.Wastage.RightSizing.Recommended.EBSBandwidth
		//iopsProperty.Recommended = i.Wastage.RightSizing.Recommended.EBSIops
		//netThroughputProperty.Recommended = i.Wastage.RightSizing.Recommended.NetworkThroughput
		//enaProperty.Recommended = i.Wastage.RightSizing.Recommended.ENASupported
	}
	ec2Instance.Properties = append(ec2Instance.Properties, regionProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, instanceSizeProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, view.Property{
		Key: style.Bold.Render("Compute"),
	})
	ec2Instance.Properties = append(ec2Instance.Properties, vCPUProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, processorProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, architectureProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, licenseCostProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, memoryProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, ebsProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, iopsProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, view.Property{
		Key: style.Bold.Render("Network Performance"),
	})
	ec2Instance.Properties = append(ec2Instance.Properties, netThroughputProperty)
	ec2Instance.Properties = append(ec2Instance.Properties, enaProperty)

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
