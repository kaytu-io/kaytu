package preferences

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"sync"
)

type PreferenceItem struct {
	Service        string   `yaml:"service"`
	Key            string   `yaml:"key"`
	Alias          string   `yaml:"alias"`
	IsNumber       bool     `yaml:"isNumber"`
	Value          *string  `yaml:"value"`
	PossibleValues []string `yaml:"possibleValues"`
	Pinned         bool     `yaml:"pinned"`
	PreventPinning bool     `yaml:"preventPinning"`
	Unit           string   `yaml:"unit"`
}

var (
	defaultPref     []PreferenceItem
	defaultPrefSync sync.Once
)

func Update(pis []PreferenceItem) {
	for _, pi := range pis {
		for idx, pref := range defaultPref {
			if pref.Service == pi.Service && pref.Key == pi.Key {
				defaultPref[idx] = pi
				break
			}
		}
	}
}

func DefaultPreferences() []PreferenceItem {
	defaultPrefSync.Do(func() {
		defaultPref = []PreferenceItem{
			{Service: "EC2Instance", Key: "Tenancy", Pinned: true, PossibleValues: []string{"", "Host", "Shared", "Dedicated"}},
			{Service: "EC2Instance", Key: "EBSOptimized", PossibleValues: []string{"", "Yes", "No"}},
			{Service: "EC2Instance", Key: "LicenseModel", PossibleValues: []string{"", "Bring your own license", "No License required"}},
			{Service: "EC2Instance", Key: "Region", Pinned: true},
			{Service: "EC2Instance", Key: "CurrentGeneration", PossibleValues: []string{"", "Yes", "No"}},
			{Service: "EC2Instance", Key: "PhysicalProcessor"},
			{Service: "EC2Instance", Key: "ClockSpeed"},
			{Service: "EC2Instance", Key: "OperatingSystem", Pinned: true, PossibleValues: []string{"", "Windows", "Linux/UNIX"}},
			{Service: "EC2Instance", Key: "ProcessorArchitecture", Pinned: true, PossibleValues: []string{"", "x86_64", "arm64", "arm64_mac"}},
			{Service: "EC2Instance", Key: "InstanceFamily", PossibleValues: []string{"", "General purpose", "Compute optimized", "Memory optimized", "Storage optimized", "FPGA Instances", "GPU instance", "Machine Learning ASIC Instances", "Media Accelerator Instances"}},
			{Service: "EC2Instance", Key: "ENASupported"},
			{Service: "EC2Instance", Key: "SupportedRootDeviceTypes", Value: aws.String("EBSOnly"), PreventPinning: true, PossibleValues: []string{"EBSOnly"}},
			{Service: "EC2Instance", Key: "vCPU", IsNumber: true},
			{Service: "EC2Instance", Key: "MemoryGB", Alias: "Memory", IsNumber: true, Pinned: true, Unit: "GiB"},
			{Service: "EC2Instance", Key: "CPUBreathingRoom", IsNumber: true, Value: aws.String("10"), PreventPinning: true, Unit: "%"},
			{Service: "EC2Instance", Key: "MemoryBreathingRoom", IsNumber: true, Value: aws.String("10"), PreventPinning: true, Unit: "%"},
			{Service: "EC2Instance", Key: "NetworkBreathingRoom", IsNumber: true, Value: aws.String("10"), PreventPinning: true, Unit: "%"},
			{Service: "EC2Instance", Key: "ObservabilityTimePeriod", Value: aws.String("7"), PreventPinning: true, Unit: "days", PossibleValues: []string{"7"}},
			{Service: "EC2Instance", Key: "RuntimeInterval", Value: aws.String("730"), PreventPinning: true, Unit: "hours", PossibleValues: []string{"730"}},
			{Service: "EC2Instance", Key: "ExcludeBurstableInstances", Value: aws.String("No"), PreventPinning: true, PossibleValues: []string{"No", "Yes"}},
			{Service: "EBSVolume", Key: "IOPS", IsNumber: true},
			{Service: "EBSVolume", Key: "Throughput", IsNumber: true, Unit: "Mbps"},
			{Service: "EBSVolume", Key: "Size", IsNumber: true, Pinned: true, Unit: "GB"},
			{Service: "EBSVolume", Key: "VolumeFamily", PossibleValues: []string{"", "General Purpose", "Solid State Drive", "IO Optimized", "Hard Disk Drive"}},
			{Service: "EBSVolume", Key: "VolumeType", PossibleValues: []string{"", "standard", "io1", "io2", "gp2", "gp3", "sc1", "st1"}},
			{Service: "EBSVolume", Key: "IOPSBreathingRoom", IsNumber: true, Value: aws.String("10"), PreventPinning: true, Unit: "%"},
			{Service: "EBSVolume", Key: "ThroughputBreathingRoom", IsNumber: true, Value: aws.String("10"), PreventPinning: true, Unit: "%"},
		}
	})
	return defaultPref
}

func Export(pref []PreferenceItem) map[string]*string {
	ex := map[string]*string{}
	for _, p := range pref {
		if p.Pinned {
			ex[p.Key] = nil
		} else {
			if p.Value != nil {
				ex[p.Key] = p.Value
			}
		}
	}
	return ex
}
