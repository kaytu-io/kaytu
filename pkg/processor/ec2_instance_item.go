package processor

import (
	"fmt"
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	preferences2 "github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/cmd/optimize/style"
	"github.com/kaytu-io/kaytu/cmd/optimize/view"
	"github.com/kaytu-io/kaytu/pkg/api/wastage"
	"github.com/kaytu-io/kaytu/pkg/hash"
	"strings"
)

type EC2InstanceItem struct {
	Instance            types.Instance
	Region              string
	OptimizationLoading bool
	Preferences         []preferences2.PreferenceItem
	Skipped             bool
	SkipReason          *string
	Volumes             []types.Volume
	Metrics             map[string][]types2.Datapoint
	VolumeMetrics       map[string]map[string][]types2.Datapoint
	Wastage             wastage.EC2InstanceWastageResponse
}

func (i EC2InstanceItem) EC2InstanceDevice() view.Device {
	ec2Instance := view.Device{
		Properties:   nil,
		DeviceID:     *i.Instance.InstanceId,
		ResourceType: "EC2 Instance",
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
		Average: view.Percentage(i.Wastage.RightSizing.VCPU.Avg),
		Max:     view.Percentage(i.Wastage.RightSizing.VCPU.Max),
	}
	processorProperty := view.Property{
		Key:     "  Processor(s)",
		Current: i.Wastage.RightSizing.Current.Processor,
	}
	architectureProperty := view.Property{
		Key:     "  Architecture",
		Current: i.Wastage.RightSizing.Current.Architecture,
	}
	licenseCostProperty := view.Property{
		Key:     "  License Cost",
		Current: fmt.Sprintf("$%.2f", i.Wastage.RightSizing.Current.LicensePrice),
	}
	memoryProperty := view.Property{
		Key:     "  Memory",
		Current: fmt.Sprintf("%.1f GiB", i.Wastage.RightSizing.Current.Memory),
		Average: view.Percentage(i.Wastage.RightSizing.Memory.Avg),
		Max:     view.Percentage(i.Wastage.RightSizing.Memory.Max),
	}
	ebsProperty := view.Property{
		Key:     "EBS Bandwidth",
		Current: fmt.Sprintf("%s", i.Wastage.RightSizing.Current.EBSBandwidth),
		Average: view.PNetworkThroughputMbps(i.Wastage.RightSizing.EBSBandwidth.Avg),
		Max:     view.PNetworkThroughputMbps(i.Wastage.RightSizing.EBSBandwidth.Max),
	}
	iopsProperty := view.Property{
		Key:     "EBS IOPS",
		Current: fmt.Sprintf("%s", i.Wastage.RightSizing.Current.EBSIops),
		Average: fmt.Sprintf("%s io/s", view.PFloat64ToString(i.Wastage.RightSizing.EBSIops.Avg)),
		Max:     fmt.Sprintf("%s io/s", view.PFloat64ToString(i.Wastage.RightSizing.EBSIops.Max)),
	}
	netThroughputProperty := view.Property{
		Key:     "  Throughput",
		Current: fmt.Sprintf("%s", i.Wastage.RightSizing.Current.NetworkThroughput),
		Average: view.PNetworkThroughputMbps(i.Wastage.RightSizing.NetworkThroughput.Avg),
		Max:     view.PNetworkThroughputMbps(i.Wastage.RightSizing.NetworkThroughput.Max),
	}
	enaProperty := view.Property{
		Key:     "  ENA",
		Current: fmt.Sprintf("%s", i.Wastage.RightSizing.Current.ENASupported),
	}

	if i.Wastage.RightSizing.Recommended != nil {
		ec2Instance.RightSizedCost = i.Wastage.RightSizing.Recommended.Cost
		regionProperty.Recommended = i.Wastage.RightSizing.Recommended.Region
		instanceSizeProperty.Recommended = i.Wastage.RightSizing.Recommended.InstanceType
		vCPUProperty.Recommended = fmt.Sprintf("%d", i.Wastage.RightSizing.Recommended.VCPU)
		processorProperty.Recommended = i.Wastage.RightSizing.Recommended.Processor
		architectureProperty.Recommended = i.Wastage.RightSizing.Recommended.Architecture
		licenseCostProperty.Recommended = fmt.Sprintf("$%.2f", i.Wastage.RightSizing.Recommended.LicensePrice)
		memoryProperty.Recommended = fmt.Sprintf("%.1f GiB", i.Wastage.RightSizing.Recommended.Memory)
		ebsProperty.Recommended = i.Wastage.RightSizing.Recommended.EBSBandwidth
		iopsProperty.Recommended = i.Wastage.RightSizing.Recommended.EBSIops
		netThroughputProperty.Recommended = i.Wastage.RightSizing.Recommended.NetworkThroughput
		enaProperty.Recommended = i.Wastage.RightSizing.Recommended.ENASupported
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

func (i EC2InstanceItem) EBSVolumeDevice(v types.Volume, vs wastage.EBSVolumeRecommendation) view.Device {
	volume := view.Device{
		Properties:   nil,
		DeviceID:     *v.VolumeId,
		ResourceType: "EBS Volume",
		Runtime:      "730 hours",
		CurrentCost:  vs.Current.Cost,
	}
	storageTierProp := view.Property{
		Key:     "  EBS Storage Tier",
		Current: string(vs.Current.Tier),
	}
	volumeSizeProp := view.Property{
		Key:     "  Volume Size (GB)",
		Current: view.SizeByteToGB(vs.Current.VolumeSize),
	}
	iopsProp := view.Property{
		Key:     style.Bold.Render("IOPS"),
		Current: fmt.Sprintf("%d", vs.Current.IOPS()),
		Average: view.PFloat64ToString(vs.IOPS.Avg),
		Max:     view.PFloat64ToString(vs.IOPS.Max),
	}
	baselineIOPSProp := view.Property{
		Key:     "  Baseline IOPS",
		Current: fmt.Sprintf("%d", vs.Current.BaselineIOPS),
	}
	provisionedIOPSProp := view.Property{
		Key:     "  Provisioned IOPS",
		Current: view.PInt32ToString(vs.Current.ProvisionedIOPS),
	}
	throughputProp := view.Property{
		Key:     style.Bold.Render("Throughput (MB/s)"),
		Current: fmt.Sprintf("%.2f", vs.Current.Throughput()),
		Average: view.PNetworkThroughputMbps(vs.Throughput.Avg),
		Max:     view.PNetworkThroughputMbps(vs.Throughput.Max),
	}
	baselineThroughputProp := view.Property{
		Key:     "  Baseline Throughput",
		Current: view.NetworkThroughputMbps(vs.Current.BaselineThroughput),
	}
	provisionedThroughputProp := view.Property{
		Key:     "  Provisioned Throughput",
		Current: view.PNetworkThroughputMbps(vs.Current.ProvisionedThroughput),
	}

	if vs.Recommended != nil {
		volume.RightSizedCost = vs.Recommended.Cost
		storageTierProp.Recommended = string(vs.Recommended.Tier)
		volumeSizeProp.Recommended = view.SizeByteToGB(vs.Recommended.VolumeSize)
		iopsProp.Recommended = fmt.Sprintf("%d", vs.Recommended.IOPS())
		baselineIOPSProp.Recommended = fmt.Sprintf("%d", vs.Recommended.BaselineIOPS)
		provisionedIOPSProp.Recommended = view.PInt32ToString(vs.Recommended.ProvisionedIOPS)
		throughputProp.Recommended = fmt.Sprintf("%.2f", vs.Recommended.Throughput())
		baselineThroughputProp.Recommended = view.NetworkThroughputMbps(vs.Recommended.BaselineThroughput)
		provisionedThroughputProp.Recommended = view.PNetworkThroughputMbps(vs.Recommended.ProvisionedThroughput)
	}

	volume.Properties = append(volume.Properties, storageTierProp)
	volume.Properties = append(volume.Properties, volumeSizeProp)
	volume.Properties = append(volume.Properties, iopsProp)
	volume.Properties = append(volume.Properties, baselineIOPSProp)
	volume.Properties = append(volume.Properties, provisionedIOPSProp)
	volume.Properties = append(volume.Properties, throughputProp)
	volume.Properties = append(volume.Properties, baselineThroughputProp)
	volume.Properties = append(volume.Properties, provisionedThroughputProp)
	return volume
}

func (i EC2InstanceItem) Devices() []view.Device {
	var devices []view.Device
	devices = append(devices, i.EC2InstanceDevice())
	for _, v := range i.Volumes {
		vs, ok := i.Wastage.VolumeRightSizing[hash.HashString(*v.VolumeId)]
		if !ok {
			continue
		}

		devices = append(devices, i.EBSVolumeDevice(v, vs))
	}
	return devices
}

func (i EC2InstanceItem) ToOptimizationItem() view.OptimizationItem {
	oi := view.OptimizationItem{
		ID:           *i.Instance.InstanceId,
		ResourceType: string(i.Instance.InstanceType),
		Region:       i.Region,
		Devices:      i.Devices(),
		Preferences:  i.Preferences,
		Description:  i.Wastage.RightSizing.Description,
		Loading:      i.OptimizationLoading,
		Skipped:      i.Skipped,
		SkipReason:   i.SkipReason,
	}

	if i.Instance.PlatformDetails != nil {
		oi.Platform = *i.Instance.PlatformDetails
	}
	for _, t := range i.Instance.Tags {
		if t.Key != nil && strings.ToLower(*t.Key) == "name" && t.Value != nil {
			oi.Name = *t.Value
		}
	}
	if oi.Name == "" {
		oi.Name = *i.Instance.InstanceId
	}

	return oi
}
