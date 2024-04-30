package view

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	preferences2 "github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/pkg/api/wastage"
	"github.com/kaytu-io/kaytu/pkg/hash"
	"os"
	"strings"
	"sync"
	"time"
)

type App struct {
	processInstanceChan chan OptimizationItem
	optimizationsTable  *Ec2InstanceOptimizations
	jobs                *JobsView

	width  int
	height int
}

var (
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"})
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func NewApp(cfg aws.Config, identification map[string]string) *App {
	pi := make(chan OptimizationItem, 1000)
	r := &App{
		processInstanceChan: pi,
		optimizationsTable:  NewEC2InstanceOptimizations(pi),
		jobs:                NewJobsView(),
	}
	go r.ProcessInstances(cfg, identification)
	go r.ProcessAllRegions(cfg)
	return r
}

func (m *App) Init() tea.Cmd {
	optTableCmd := m.optimizationsTable.Init()
	return tea.Batch(optTableCmd, tea.EnterAltScreen)
}

func (m *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.UpdateResponsive()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	m.jobs.Update(msg)
	_, optTableCmd := m.optimizationsTable.Update(msg)
	return m, tea.Batch(optTableCmd)
}

func (m *App) View() string {
	if !m.checkResponsive() {
		return "Application cannot be rendered in this screen size, please increase height of your terminal"
	}
	sb := strings.Builder{}
	sb.WriteString(m.optimizationsTable.View())
	sb.WriteString(m.jobs.View())
	return sb.String()
}

func (m *App) checkResponsive() bool {
	return m.height >= m.jobs.height+m.optimizationsTable.height && m.jobs.IsResponsive() && m.optimizationsTable.IsResponsive()
}

func (m *App) UpdateResponsive() {
	m.optimizationsTable.SetHeight(m.optimizationsTable.MinHeight())
	m.jobs.SetHeight(m.jobs.MinHeight())

	if !m.checkResponsive() {
		return // nothing we can do
	}

	for m.optimizationsTable.height < m.optimizationsTable.PreferredMinHeight() {
		m.optimizationsTable.SetHeight(m.optimizationsTable.height + 1)
		if !m.checkResponsive() {
			m.optimizationsTable.SetHeight(m.optimizationsTable.height - 1)
			return
		}
	}

	for m.jobs.height < m.jobs.MaxHeight() {
		m.jobs.SetHeight(m.jobs.height + 1)
		if !m.checkResponsive() {
			m.jobs.SetHeight(m.jobs.height - 1)
			return
		}
	}

	for m.optimizationsTable.height < m.optimizationsTable.MaxHeight() {
		m.optimizationsTable.SetHeight(m.optimizationsTable.height + 1)
		if !m.checkResponsive() {
			m.optimizationsTable.SetHeight(m.optimizationsTable.height - 1)
			return
		}
	}
}

func (m *App) ProcessInstances(awsCfg aws.Config, identification map[string]string) {
	for item := range m.processInstanceChan {
		awsCfg.Region = item.Region
		localAWSCfg := awsCfg
		localItem := item

		go m.ProcessInstance(localAWSCfg, localItem, identification)
	}
}

func (m *App) ProcessInstance(awsConf aws.Config, item OptimizationItem, identification map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			m.jobs.PublishError(fmt.Errorf("%v", r))
		}
	}()

	client := ec2.NewFromConfig(awsConf)
	var volumeIds []string
	for _, bd := range item.Instance.BlockDeviceMappings {
		if bd.Ebs == nil {
			continue
		}
		volumeIds = append(volumeIds, *bd.Ebs.VolumeId)
	}

	job := m.jobs.Publish(Job{ID: fmt.Sprintf("volumes_%s", *item.Instance.InstanceId), Descrption: fmt.Sprintf("getting volumes of %s", *item.Instance.InstanceId)})
	job.Done = true

	volumesResp, err := client.DescribeVolumes(context.Background(), &ec2.DescribeVolumesInput{
		VolumeIds: volumeIds,
	})
	if err != nil {
		job.FailureMessage = err.Error()
		m.jobs.Publish(job)
		return
	}
	m.jobs.Publish(job)

	req, err := m.getEc2InstanceRequestData(context.Background(), awsConf, item.Instance, volumesResp.Volumes, preferences2.Export(item.Preferences), identification)
	if err != nil {
		job.FailureMessage = err.Error()
		m.jobs.Publish(job)
		return
	}

	job = Job{ID: fmt.Sprintf("wastage_%s", *item.Instance.InstanceId), Descrption: fmt.Sprintf("Evaluating usage data for %s", *item.Instance.InstanceId)}
	m.jobs.Publish(job)
	job.Done = true

	res, err := wastage.Ec2InstanceWastageRequest(*req)
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
		m.optimizationsTable.SendItem(item)
		return
	}

	m.optimizationsTable.SendItem(OptimizationItem{
		Instance:            item.Instance,
		Volumes:             volumesResp.Volumes,
		Region:              awsConf.Region,
		OptimizationLoading: false,
		Wastage:             *res,
		Preferences:         item.Preferences,
	})
}

func (m *App) ProcessRegion(cfg aws.Config) {
	ctx := context.Background()
	defer func() {
		if r := recover(); r != nil {
			m.jobs.PublishError(fmt.Errorf("%v", r))
		}
	}()
	client := ec2.NewFromConfig(cfg)

	job := Job{ID: fmt.Sprintf("region_ec2_instances_%s", cfg.Region), Descrption: "Listing all ec2 instances in " + cfg.Region}
	m.jobs.Publish(job)
	job.Done = true
	defer func() {
		m.jobs.Publish(job)
	}()

	paginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			job.FailureMessage = err.Error()
			return
		}

		for _, r := range page.Reservations {
			for _, v := range r.Instances {
				isAutoScaling := false
				for _, tag := range v.Tags {
					if *tag.Key == "aws:autoscaling:groupName" && tag.Value != nil && *tag.Value != "" {
						isAutoScaling = true
					}
				}

				preferences := preferences2.DefaultPreferences()
				oi := OptimizationItem{
					Instance:            v,
					Region:              cfg.Region,
					OptimizationLoading: true,
					Preferences:         preferences,
				}
				if v.State.Name != types.InstanceStateNameRunning ||
					v.InstanceLifecycle == types.InstanceLifecycleTypeSpot ||
					isAutoScaling {
					oi.OptimizationLoading = false
					oi.Skipped = true
					reason := ""
					if v.State.Name != types.InstanceStateNameRunning {
						reason = "not running"
					} else if v.InstanceLifecycle == types.InstanceLifecycleTypeSpot {
						reason = "spot instance"
					} else if isAutoScaling {
						reason = "auto-scaling group instance"
					}
					oi.SkipReason = &reason
				}
				m.optimizationsTable.SendItem(oi)
				if !oi.Skipped {
					m.processInstanceChan <- oi
				}
			}
		}
	}
}

func (m *App) ProcessAllRegions(cfg aws.Config) {
	defer func() {
		if r := recover(); r != nil {
			m.jobs.PublishError(fmt.Errorf("%v", r))
		}
	}()
	regionClient := ec2.NewFromConfig(cfg)

	job := m.jobs.Publish(Job{ID: "list_all_regions", Descrption: "Listing all available regions"})
	job.Done = true
	regions, err := regionClient.DescribeRegions(context.Background(), &ec2.DescribeRegionsInput{AllRegions: aws.Bool(false)})
	if err != nil {
		job.FailureMessage = err.Error()
		m.jobs.Publish(job)
		return
	}
	m.jobs.Publish(job)

	wg := sync.WaitGroup{}
	wg.Add(len(regions.Regions))

	for _, region := range regions.Regions {
		localCfg := cfg
		localCfg.Region = *region.RegionName

		go func() {
			defer wg.Done()
			m.ProcessRegion(localCfg)
		}()
	}
	wg.Wait()
}

func (m *App) getEc2InstanceRequestData(ctx context.Context, cfg aws.Config, instance types.Instance, volumes []types.Volume, preferences map[string]*string, identification map[string]string) (*wastage.EC2InstanceWastageRequest, error) {
	cloudwatchClient := cloudwatch.NewFromConfig(cfg)
	startTime := time.Now().Add(-24 * 7 * time.Hour)
	endTime := time.Now()
	statistics := []types2.Statistic{
		types2.StatisticAverage,
		types2.StatisticMinimum,
		types2.StatisticMaximum,
	}
	dimensionFilter := []types2.Dimension{
		{
			Name:  aws.String("InstanceId"),
			Value: instance.InstanceId,
		},
	}
	metrics := map[string][]types2.Datapoint{}

	job := m.jobs.Publish(Job{ID: fmt.Sprintf("metrics_%s", *instance.InstanceId), Descrption: fmt.Sprintf("Gathering monitoring metrics for %s", *instance.InstanceId)})
	job.Done = true

	paginator := cloudwatch.NewListMetricsPaginator(cloudwatchClient, &cloudwatch.ListMetricsInput{
		Namespace: aws.String("AWS/EC2"),
		Dimensions: []types2.DimensionFilter{
			{
				Name:  aws.String("InstanceId"),
				Value: instance.InstanceId,
			},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			job.FailureMessage = err.Error()
			m.jobs.Publish(job)
			return nil, err
		}

		for _, p := range page.Metrics {
			if p.MetricName == nil || (*p.MetricName != "CPUUtilization" &&
				*p.MetricName != "NetworkIn" &&
				*p.MetricName != "NetworkOut") {
				continue
			}

			// Create input for GetMetricStatistics
			input := &cloudwatch.GetMetricStatisticsInput{
				Namespace:  aws.String("AWS/EC2"),
				MetricName: p.MetricName,
				Dimensions: dimensionFilter,
				StartTime:  aws.Time(startTime),
				EndTime:    aws.Time(endTime),
				Period:     aws.Int32(60 * 60), // 1 hour intervals
				Statistics: statistics,
			}

			// Get metric data
			resp, err := cloudwatchClient.GetMetricStatistics(ctx, input)
			if err != nil {
				job.FailureMessage = err.Error()
				m.jobs.Publish(job)
				return nil, err
			}

			metrics[*p.MetricName] = resp.Datapoints
		}
	}
	m.jobs.Publish(job)

	job = Job{ID: fmt.Sprintf("metrics_cw_%s", *instance.InstanceId), Descrption: fmt.Sprintf("getting cloud watch agent metrics of %s", *instance.InstanceId)}
	m.jobs.Publish(job)
	job.Done = true

	paginator = cloudwatch.NewListMetricsPaginator(cloudwatchClient, &cloudwatch.ListMetricsInput{
		Namespace: aws.String("CWAgent"),
		Dimensions: []types2.DimensionFilter{
			{
				Name:  aws.String("InstanceId"),
				Value: instance.InstanceId,
			},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			job.FailureMessage = err.Error()
			m.jobs.Publish(job)
			return nil, err
		}

		for _, p := range page.Metrics {
			if p.MetricName == nil || (*p.MetricName != "mem_used_percent") {
				continue
			}

			// Create input for GetMetricStatistics
			input := &cloudwatch.GetMetricStatisticsInput{
				Namespace:  aws.String("CWAgent"),
				MetricName: p.MetricName,
				Dimensions: dimensionFilter,
				StartTime:  aws.Time(startTime),
				EndTime:    aws.Time(endTime),
				Period:     aws.Int32(60 * 60), // 1 hour intervals
				Statistics: statistics,
			}

			// Get metric data
			resp, err := cloudwatchClient.GetMetricStatistics(ctx, input)
			if err != nil {
				job.FailureMessage = err.Error()
				m.jobs.Publish(job)
				return nil, err
			}

			metrics[*p.MetricName] = resp.Datapoints
		}
	}
	m.jobs.Publish(job)

	var monitoring *types.MonitoringState
	if instance.Monitoring != nil {
		monitoring = &instance.Monitoring.State
	}
	var placement *wastage.EC2Placement
	if instance.Placement != nil {
		placement = &wastage.EC2Placement{
			Tenancy: instance.Placement.Tenancy,
		}
		if instance.Placement.AvailabilityZone != nil {
			placement.AvailabilityZone = *instance.Placement.AvailabilityZone
		}
		if instance.Placement.HostId != nil {
			placement.HashedHostId = hash.HashString(*instance.Placement.HostId)
		}
	}

	var kaytuVolumes []wastage.EC2Volume
	volumeMetrics := map[string]map[string][]types2.Datapoint{}
	for _, v := range volumes {
		kaytuVolumes = append(kaytuVolumes, toEBSVolume(v))

		job = Job{ID: fmt.Sprintf("metrics_volume_%s", *instance.InstanceId), Descrption: fmt.Sprintf("getting volume metrics of %s", *v.VolumeId)}
		m.jobs.Publish(job)
		job.Done = true

		paginator := cloudwatch.NewListMetricsPaginator(cloudwatchClient, &cloudwatch.ListMetricsInput{
			Namespace: aws.String("AWS/EBS"),
			Dimensions: []types2.DimensionFilter{
				{
					Name:  aws.String("VolumeId"),
					Value: v.VolumeId,
				},
			},
		})
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				job.FailureMessage = err.Error()
				m.jobs.Publish(job)
				return nil, err
			}

			for _, p := range page.Metrics {
				if p.MetricName == nil || (*p.MetricName != "VolumeReadOps" &&
					*p.MetricName != "VolumeWriteOps" &&
					*p.MetricName != "VolumeReadBytes" &&
					*p.MetricName != "VolumeWriteBytes") {
					continue
				}

				// Create input for GetMetricStatistics
				input := &cloudwatch.GetMetricStatisticsInput{
					Namespace:  aws.String("AWS/EBS"),
					MetricName: p.MetricName,
					Dimensions: []types2.Dimension{
						{
							Name:  aws.String("VolumeId"),
							Value: v.VolumeId,
						},
					},
					StartTime:  aws.Time(startTime),
					EndTime:    aws.Time(endTime),
					Period:     aws.Int32(60 * 60), // 1 hour intervals
					Statistics: statistics,
				}

				// Get metric data
				resp, err := cloudwatchClient.GetMetricStatistics(ctx, input)
				if err != nil {
					job.FailureMessage = err.Error()
					m.jobs.Publish(job)
					return nil, err
				}

				if _, ok := volumeMetrics[hash.HashString(*v.VolumeId)]; !ok {
					volumeMetrics[hash.HashString(*v.VolumeId)] = make(map[string][]types2.Datapoint)
				}
				volumeMetrics[hash.HashString(*v.VolumeId)][*p.MetricName] = resp.Datapoints
			}
		}

		m.jobs.Publish(job)
	}
	platform := ""
	if instance.PlatformDetails != nil {
		platform = *instance.PlatformDetails
	}

	return &wastage.EC2InstanceWastageRequest{
		Identification: identification,
		Instance: wastage.EC2Instance{
			HashedInstanceId:  hash.HashString(*instance.InstanceId),
			State:             instance.State.Name,
			InstanceType:      instance.InstanceType,
			Platform:          platform,
			ThreadsPerCore:    *instance.CpuOptions.ThreadsPerCore,
			CoreCount:         *instance.CpuOptions.CoreCount,
			EbsOptimized:      *instance.EbsOptimized,
			InstanceLifecycle: instance.InstanceLifecycle,
			Monitoring:        monitoring,
			Placement:         placement,
			UsageOperation:    *instance.UsageOperation,
			Tenancy:           instance.Placement.Tenancy,
		},
		Volumes:       kaytuVolumes,
		Metrics:       metrics,
		VolumeMetrics: volumeMetrics,
		Region:        cfg.Region,
		Preferences:   preferences,
	}, nil
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
