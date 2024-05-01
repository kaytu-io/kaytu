package processor

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	preferences2 "github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/cmd/optimize/view"
	"github.com/kaytu-io/kaytu/pkg/api/wastage"
	"github.com/kaytu-io/kaytu/pkg/hash"
	"github.com/kaytu-io/kaytu/pkg/metrics"
	"github.com/kaytu-io/kaytu/pkg/provider"
	"os"
	"strings"
	"sync"
	"time"
)

type EC2InstanceProcessor struct {
	jobs              *view.JobsView
	optimizationsView *view.OptimizationsView

	provider       provider.Provider
	metricProvider metrics.MetricProvider
	identification map[string]string

	processWastageChan chan EC2InstanceItem
	items              map[string]EC2InstanceItem
}

func NewEC2InstanceProcessor(
	prv provider.Provider,
	metric metrics.MetricProvider,
	identification map[string]string,
	jobs *view.JobsView,
	optimizationsView *view.OptimizationsView,
) *EC2InstanceProcessor {
	r := &EC2InstanceProcessor{
		processWastageChan: make(chan EC2InstanceItem, 1000),
		optimizationsView:  optimizationsView,
		jobs:               jobs,
		provider:           prv,
		metricProvider:     metric,
		identification:     identification,
		items:              map[string]EC2InstanceItem{},
	}
	go r.ProcessWastages()
	go r.ProcessAllRegions()
	return r
}

func (m *EC2InstanceProcessor) ProcessAllRegions() {
	defer func() {
		if r := recover(); r != nil {
			m.jobs.PublishError(fmt.Errorf("%v", r))
		}
	}()

	job := m.jobs.Publish(view.Job{ID: "list_all_regions", Descrption: "Listing all available regions"})
	job.Done = true
	regions, err := m.provider.ListAllRegions()
	if err != nil {
		job.FailureMessage = err.Error()
		m.jobs.Publish(job)
		return
	}
	m.jobs.Publish(job)

	wg := sync.WaitGroup{}
	wg.Add(len(regions))
	for _, region := range regions {
		region := region
		go func() {
			defer wg.Done()
			m.ProcessRegion(region)
		}()
	}
	wg.Wait()
}

func (m *EC2InstanceProcessor) ProcessRegion(region string) {
	defer func() {
		if r := recover(); r != nil {
			m.jobs.PublishError(fmt.Errorf("%v", r))
		}
	}()

	job := m.jobs.Publish(view.Job{ID: fmt.Sprintf("region_ec2_instances_%s", region), Descrption: "Listing all ec2 instances in " + region})
	job.Done = true

	instances, err := m.provider.ListInstances(region)
	if err != nil {
		job.FailureMessage = err.Error()
		m.jobs.Publish(job)
		return
	}
	m.jobs.Publish(job)

	for _, instance := range instances {
		oi := EC2InstanceItem{
			Instance:            instance,
			Region:              region,
			OptimizationLoading: true,
			Preferences:         preferences2.DefaultPreferences(),
		}

		isAutoScaling := false
		for _, tag := range instance.Tags {
			if *tag.Key == "aws:autoscaling:groupName" && tag.Value != nil && *tag.Value != "" {
				isAutoScaling = true
			}
		}
		if instance.State.Name != types.InstanceStateNameRunning ||
			instance.InstanceLifecycle == types.InstanceLifecycleTypeSpot ||
			isAutoScaling {
			oi.OptimizationLoading = false
			oi.Skipped = true
			reason := ""
			if instance.State.Name != types.InstanceStateNameRunning {
				reason = "not running"
			} else if instance.InstanceLifecycle == types.InstanceLifecycleTypeSpot {
				reason = "spot instance"
			} else if isAutoScaling {
				reason = "auto-scaling group instance"
			}
			if len(reason) > 0 {
				oi.SkipReason = &reason
			}
		}

		// just to show the loading
		m.items[*oi.Instance.InstanceId] = oi
		m.optimizationsView.SendItem(oi.ToOptimizationItem())
	}

	for _, instance := range instances {
		isAutoScaling := false
		for _, tag := range instance.Tags {
			if *tag.Key == "aws:autoscaling:groupName" && tag.Value != nil && *tag.Value != "" {
				isAutoScaling = true
			}
		}
		if instance.State.Name != types.InstanceStateNameRunning ||
			instance.InstanceLifecycle == types.InstanceLifecycleTypeSpot ||
			isAutoScaling {
			continue
		}

		vjob := m.jobs.Publish(view.Job{ID: fmt.Sprintf("volumes_%s", *instance.InstanceId), Descrption: fmt.Sprintf("getting volumes of %s", *instance.InstanceId)})
		vjob.Done = true

		volumes, err := m.provider.ListAttachedVolumes(region, instance)
		if err != nil {
			vjob.FailureMessage = err.Error()
			m.jobs.Publish(vjob)
			return
		}
		m.jobs.Publish(vjob)

		imjob := m.jobs.Publish(view.Job{ID: fmt.Sprintf("instance_%s_metrics", *instance.InstanceId), Descrption: fmt.Sprintf("getting metrics of %s", *instance.InstanceId)})
		imjob.Done = true
		startTime := time.Now().Add(-24 * 7 * time.Hour)
		endTime := time.Now()
		instanceMetrics := map[string][]types2.Datapoint{}
		cwMetrics, err := m.metricProvider.GetMetrics(
			region,
			"AWS/EC2",
			[]string{
				"CPUUtilization",
				"NetworkIn",
				"NetworkOut",
			},
			map[string][]string{
				"InstanceId": {*instance.InstanceId},
			},
			startTime, endTime,
			time.Hour,
			[]types2.Statistic{
				types2.StatisticAverage,
				types2.StatisticMaximum,
			},
		)
		if err != nil {
			imjob.FailureMessage = err.Error()
			m.jobs.Publish(imjob)
			return
		}
		for k, v := range cwMetrics {
			instanceMetrics[k] = v
		}

		cwaMetrics, err := m.metricProvider.GetMetrics(
			region,
			"CWAgent",
			[]string{
				"mem_used_percent",
			},
			map[string][]string{
				"InstanceId": {*instance.InstanceId},
			},
			startTime, endTime,
			time.Hour,
			[]types2.Statistic{
				types2.StatisticAverage,
				types2.StatisticMaximum,
			},
		)
		if err != nil {
			imjob.FailureMessage = err.Error()
			m.jobs.Publish(imjob)
			return
		}
		for k, v := range cwaMetrics {
			instanceMetrics[k] = v
		}
		m.jobs.Publish(imjob)

		ivjob := m.jobs.Publish(view.Job{ID: fmt.Sprintf("volume_%s_metrics", *instance.InstanceId), Descrption: fmt.Sprintf("getting volume metrics of %s", *instance.InstanceId)})
		ivjob.Done = true

		var volumeIDs []string
		for _, v := range instance.BlockDeviceMappings {
			if v.Ebs != nil {
				volumeIDs = append(volumeIDs, *v.Ebs.VolumeId)
			}
		}

		volumeMetrics := map[string]map[string][]types2.Datapoint{}
		for _, v := range volumeIDs {
			volumeMetric, err := m.metricProvider.GetMetrics(
				region,
				"AWS/EBS",
				[]string{
					"VolumeReadOps",
					"VolumeWriteOps",
					"VolumeReadBytes",
					"VolumeWriteBytes",
				},
				map[string][]string{
					"VolumeId": {v},
				},
				startTime, endTime,
				time.Hour,
				[]types2.Statistic{
					types2.StatisticAverage,
					types2.StatisticMaximum,
				},
			)
			if err != nil {
				ivjob.FailureMessage = err.Error()
				m.jobs.Publish(ivjob)
				return
			}
			volumeMetrics[v] = volumeMetric
		}
		m.jobs.Publish(ivjob)

		oi := EC2InstanceItem{
			Instance:            instance,
			Volumes:             volumes,
			Metrics:             instanceMetrics,
			VolumeMetrics:       volumeMetrics,
			Region:              region,
			OptimizationLoading: true,
			Preferences:         preferences2.DefaultPreferences(),
		}
		if instance.State.Name != types.InstanceStateNameRunning ||
			instance.InstanceLifecycle == types.InstanceLifecycleTypeSpot ||
			isAutoScaling {
			oi.OptimizationLoading = false
			oi.Skipped = true
			reason := ""
			if instance.State.Name != types.InstanceStateNameRunning {
				reason = "not running"
			} else if instance.InstanceLifecycle == types.InstanceLifecycleTypeSpot {
				reason = "spot instance"
			} else if isAutoScaling {
				reason = "auto-scaling group instance"
			}
			if len(reason) > 0 {
				oi.SkipReason = &reason
			}
		}
		m.items[*oi.Instance.InstanceId] = oi
		m.optimizationsView.SendItem(oi.ToOptimizationItem())
		if !oi.Skipped {
			m.processWastageChan <- oi
		}
	}
}

func (m *EC2InstanceProcessor) ProcessWastages() {
	for item := range m.processWastageChan {
		go m.WastageWorker(item)
	}
}

func (m *EC2InstanceProcessor) WastageWorker(item EC2InstanceItem) {
	defer func() {
		if r := recover(); r != nil {
			m.jobs.PublishError(fmt.Errorf("%v", r))
		}
	}()

	job := m.jobs.Publish(view.Job{ID: fmt.Sprintf("wastage_%s", *item.Instance.InstanceId), Descrption: fmt.Sprintf("Evaluating usage data for %s", *item.Instance.InstanceId)})
	job.Done = true

	var monitoring *types.MonitoringState
	if item.Instance.Monitoring != nil {
		monitoring = &item.Instance.Monitoring.State
	}
	var placement *wastage.EC2Placement
	if item.Instance.Placement != nil {
		placement = &wastage.EC2Placement{
			Tenancy: item.Instance.Placement.Tenancy,
		}
		if item.Instance.Placement.AvailabilityZone != nil {
			placement.AvailabilityZone = *item.Instance.Placement.AvailabilityZone
		}
		if item.Instance.Placement.HostId != nil {
			placement.HashedHostId = hash.HashString(*item.Instance.Placement.HostId)
		}
	}
	platform := ""
	if item.Instance.PlatformDetails != nil {
		platform = *item.Instance.PlatformDetails
	}

	var volumes []wastage.EC2Volume
	for _, v := range item.Volumes {
		volumes = append(volumes, toEBSVolume(v))
	}

	res, err := wastage.Ec2InstanceWastageRequest(wastage.EC2InstanceWastageRequest{
		Identification: m.identification,
		Instance: wastage.EC2Instance{
			HashedInstanceId:  hash.HashString(*item.Instance.InstanceId),
			State:             item.Instance.State.Name,
			InstanceType:      item.Instance.InstanceType,
			Platform:          platform,
			ThreadsPerCore:    *item.Instance.CpuOptions.ThreadsPerCore,
			CoreCount:         *item.Instance.CpuOptions.CoreCount,
			EbsOptimized:      *item.Instance.EbsOptimized,
			InstanceLifecycle: item.Instance.InstanceLifecycle,
			Monitoring:        monitoring,
			Placement:         placement,
			UsageOperation:    *item.Instance.UsageOperation,
			Tenancy:           item.Instance.Placement.Tenancy,
		},
		Volumes:       volumes,
		Metrics:       item.Metrics,
		VolumeMetrics: item.VolumeMetrics,
		Region:        item.Region,
		Preferences:   preferences2.Export(item.Preferences),
	})
	if err != nil {
		if strings.Contains(err.Error(), "please login") {
			fmt.Println(err.Error())
			os.Exit(1)
			return
		}
		job.FailureMessage = err.Error()
		m.jobs.Publish(job)
		return
	}
	m.jobs.Publish(job)

	if res.RightSizing.Current.InstanceType == "" {
		item.OptimizationLoading = false
		m.items[*item.Instance.InstanceId] = item
		m.optimizationsView.SendItem(item.ToOptimizationItem())
		return
	}

	item = EC2InstanceItem{
		Instance:            item.Instance,
		Region:              item.Region,
		OptimizationLoading: false,
		Preferences:         item.Preferences,
		Skipped:             false,
		SkipReason:          nil,
		Volumes:             item.Volumes,
		Metrics:             item.Metrics,
		VolumeMetrics:       item.VolumeMetrics,
		Wastage:             *res,
	}
	m.items[*item.Instance.InstanceId] = item
	m.optimizationsView.SendItem(item.ToOptimizationItem())
}

func (m *EC2InstanceProcessor) ReEvaluate(id string, items []preferences2.PreferenceItem) {
	v := m.items[id]
	v.Preferences = items
	m.items[id] = v
	m.processWastageChan <- m.items[id]
}

func toEBSVolume(v types.Volume) wastage.EC2Volume {
	var throughput *float64
	if v.Throughput != nil {
		throughput = aws.Float64(float64(*v.Throughput))
	}

	return wastage.EC2Volume{
		HashedVolumeId:   hash.HashString(*v.VolumeId),
		VolumeType:       v.VolumeType,
		Size:             v.Size,
		Iops:             v.Iops,
		AvailabilityZone: v.AvailabilityZone,
		Throughput:       throughput,
	}
}
