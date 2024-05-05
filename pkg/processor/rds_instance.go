package processor

import (
	"fmt"
	types2 "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/google/uuid"
	preferences2 "github.com/kaytu-io/kaytu/cmd/optimize/preferences"
	"github.com/kaytu-io/kaytu/cmd/optimize/view"
	"github.com/kaytu-io/kaytu/cmd/predef"
	"github.com/kaytu-io/kaytu/pkg/api/wastage"
	"github.com/kaytu-io/kaytu/pkg/hash"
	"github.com/kaytu-io/kaytu/pkg/metrics"
	"github.com/kaytu-io/kaytu/pkg/provider"
	"os"
	"strings"
	"sync"
	"time"
)

type RDSInstanceProcessor struct {
	jobs              *view.JobsView
	optimizationsView *view.OptimizationsView

	provider       provider.Provider
	metricProvider metrics.MetricProvider
	identification map[string]string

	processWastageChan chan RDSInstanceItem
	items              map[string]RDSInstanceItem
}

func NewRDSInstanceProcessor(
	prv provider.Provider,
	metric metrics.MetricProvider,
	identification map[string]string,
	jobs *view.JobsView,
	optimizationsView *view.OptimizationsView,
) *RDSInstanceProcessor {
	r := &RDSInstanceProcessor{
		processWastageChan: make(chan RDSInstanceItem, 1000),
		optimizationsView:  optimizationsView,
		jobs:               jobs,
		provider:           prv,
		metricProvider:     metric,
		identification:     identification,
		items:              map[string]RDSInstanceItem{},
	}
	go r.ProcessWastages()
	go r.ProcessAllRegions()
	return r
}

func (m *RDSInstanceProcessor) ProcessAllRegions() {
	defer func() {
		if r := recover(); r != nil {
			m.jobs.PublishError(fmt.Errorf("%v", r))
		}
	}()

	job := m.jobs.Publish(view.Job{ID: "list_rds_all_regions", Descrption: "Listing all available regions"})
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

func (m *RDSInstanceProcessor) ProcessRegion(region string) {
	defer func() {
		if r := recover(); r != nil {
			m.jobs.PublishError(fmt.Errorf("%v", r))
		}
	}()

	job := m.jobs.Publish(view.Job{ID: fmt.Sprintf("region_rds_instances_%s", region), Descrption: "Listing all rds instances in " + region})
	job.Done = true

	instances, err := m.provider.ListRDSInstance(region)
	if err != nil {
		job.FailureMessage = err.Error()
		m.jobs.Publish(job)
		return
	}
	m.jobs.Publish(job)

	for _, instance := range instances {
		oi := RDSInstanceItem{
			Instance:            instance,
			Region:              region,
			OptimizationLoading: true,
			Preferences:         preferences2.DefaultPreferences(map[string]bool{"RDSInstance": true}),
		}

		// just to show the loading
		m.items[*oi.Instance.DBInstanceIdentifier] = oi
		m.optimizationsView.SendItem(oi.ToOptimizationItem())
	}

	for _, instance := range instances {
		imjob := m.jobs.Publish(view.Job{ID: fmt.Sprintf("rds_instance_%s_metrics", *instance.DBInstanceIdentifier), Descrption: fmt.Sprintf("getting metrics of %s", *instance.DBInstanceIdentifier)})
		imjob.Done = true
		startTime := time.Now().Add(-24 * 7 * time.Hour)
		endTime := time.Now()
		instanceMetrics := map[string][]types2.Datapoint{}
		cwMetrics, err := m.metricProvider.GetMetrics(
			region,
			"AWS/RDS",
			[]string{
				"CPUUtilization",
				"FreeableMemory",
				"FreeStorageSpace",
				"NetworkReceiveThroughput",
				"NetworkTransmitThroughput",
				"ReadIOPS",
				"ReadThroughput",
				"WriteIOPS",
				"WriteThroughput",
			},
			map[string][]string{
				"DBInstanceIdentifier": {*instance.DBInstanceIdentifier},
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
		m.jobs.Publish(imjob)

		oi := RDSInstanceItem{
			Instance:            instance,
			Metrics:             instanceMetrics,
			Region:              region,
			OptimizationLoading: true,
			Preferences:         preferences2.DefaultPreferences(map[string]bool{"RDSInstance": true}),
		}

		m.items[*oi.Instance.DBInstanceIdentifier] = oi
		m.optimizationsView.SendItem(oi.ToOptimizationItem())
		if !oi.Skipped {
			m.processWastageChan <- oi
		}
	}
}

func (m *RDSInstanceProcessor) ProcessWastages() {
	for item := range m.processWastageChan {
		go m.WastageWorker(item)
	}
}

func (m *RDSInstanceProcessor) WastageWorker(item RDSInstanceItem) {
	defer func() {
		if r := recover(); r != nil {
			m.jobs.PublishError(fmt.Errorf("%v", r))
		}
	}()

	job := m.jobs.Publish(view.Job{ID: fmt.Sprintf("wastage_rds_%s", *item.Instance.DBInstanceIdentifier), Descrption: fmt.Sprintf("Evaluating RDS usage data for %s", *item.Instance.DBInstanceIdentifier)})
	job.Done = true

	var clusterType wastage.AwsRdsClusterType
	multiAZ := item.Instance.MultiAZ != nil && *item.Instance.MultiAZ
	readableStandbys := item.Instance.ReplicaMode == types.ReplicaModeOpenReadOnly
	if multiAZ && readableStandbys {
		clusterType = wastage.AwsRdsClusterTypeMultiAzTwoInstance
	} else if multiAZ {
		clusterType = wastage.AwsRdsClusterTypeMultiAzOneInstance
	} else {
		clusterType = wastage.AwsRdsClusterTypeSingleInstance
	}

	//item.Instance.MultiAZ
	//item.Instance.ReadReplicaDBClusterIdentifiers
	//item.Instance.ReadReplicaDBInstanceIdentifiers
	//item.Instance.ReadReplicaSourceDBClusterIdentifier
	//item.Instance.ReadReplicaSourceDBInstanceIdentifier
	//item.Instance.ReplicaMode

	id := uuid.New()
	requestId := id.String()
	req := wastage.AwsRdsWastageRequest{
		RequestId:      requestId,
		CliVersion:     predef.GetVersion(),
		Identification: m.identification,
		Instance: wastage.AwsRds{
			HashedInstanceId:                   hash.HashString(*item.Instance.DBInstanceIdentifier),
			AvailabilityZone:                   *item.Instance.AvailabilityZone,
			InstanceType:                       *item.Instance.DBInstanceClass,
			Engine:                             *item.Instance.Engine,
			EngineVersion:                      *item.Instance.EngineVersion,
			LicenseModel:                       *item.Instance.LicenseModel,
			BackupRetentionPeriod:              item.Instance.BackupRetentionPeriod,
			ClusterType:                        clusterType,
			PerformanceInsightsEnabled:         *item.Instance.PerformanceInsightsEnabled,
			PerformanceInsightsRetentionPeriod: item.Instance.PerformanceInsightsRetentionPeriod,
			StorageType:                        item.Instance.StorageType,
			StorageSize:                        item.Instance.AllocatedStorage,
			StorageIops:                        item.Instance.Iops,
		},
		Metrics:     item.Metrics,
		Region:      item.Region,
		Preferences: preferences2.Export(item.Preferences, map[string]bool{"RDSInstance": true}),
	}
	if item.Instance.StorageThroughput != nil {
		floatThroughput := float64(*item.Instance.StorageThroughput)
		req.Instance.StorageThroughput = &floatThroughput
	}
	res, err := wastage.RDSInstanceWastageRequest(req)
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
		m.items[*item.Instance.DBInstanceIdentifier] = item
		m.optimizationsView.SendItem(item.ToOptimizationItem())
		return
	}

	item = RDSInstanceItem{
		Instance:            item.Instance,
		Region:              item.Region,
		OptimizationLoading: false,
		Preferences:         item.Preferences,
		Skipped:             false,
		SkipReason:          nil,
		Metrics:             item.Metrics,
		Wastage:             *res,
	}
	m.items[*item.Instance.DBInstanceIdentifier] = item
	m.optimizationsView.SendItem(item.ToOptimizationItem())
}

func (m *RDSInstanceProcessor) ReEvaluate(id string, items []preferences2.PreferenceItem) {
	v := m.items[id]
	v.Preferences = items
	m.items[id] = v
	m.processWastageChan <- m.items[id]
}
