package utils

import "fmt"

func PFloat64ToString(v *float64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f", *v)
}

func Percentage(v *float64) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%.2f%%", *v)
}

func MemoryUsagePercentageByFreeSpace(freeSpaceBytes *float64, memorySizeGB float64) string {
	if freeSpaceBytes == nil {
		return ""
	}
	memorySizeBytes := memorySizeGB * (1024 * 1024 * 1024)
	usage := memorySizeBytes - *freeSpaceBytes
	usagePercentage := usage / memorySizeBytes
	return Percentage(&usagePercentage)
}

func StorageUsagePercentageByFreeSpace(freeSpaceBytes *float64, storageSizeGB *int32) string {
	if freeSpaceBytes == nil || storageSizeGB == nil {
		return ""
	}
	storageSizeBytes := *storageSizeGB * (1024 * 1024 * 1024)
	usage := float64(storageSizeBytes) - *freeSpaceBytes
	usagePercentage := usage / float64(storageSizeBytes)
	return Percentage(&usagePercentage)
}

func PNetworkThroughputMbps(v *float64) string {
	if v == nil {
		return ""
	}
	vv := *v / (1024 * 1024) * 8
	return fmt.Sprintf("%.2f Mbps", vv)
}

func PStorageThroughputMbps(v *float64) string {
	if v == nil {
		return ""
	}
	vv := *v / (1024.0 * 1024.0) * 8.0
	return fmt.Sprintf("%.2f Mbps", vv)
}

func NetworkThroughputMbps(v float64) string {
	return fmt.Sprintf("%.2f Mbps", v/(1024.0*1024.0))
}

func PInt32ToString(v *int32) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}

func PString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func SizeByteToGB(v *int32) string {
	if v == nil {
		return ""
	}
	vv := *v // / 1000000000
	return fmt.Sprintf("%d GB", vv)
}
