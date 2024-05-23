package utils

import (
	"fmt"
	"math"
	"strings"
)

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
	usagePercentage := (usage / memorySizeBytes) * 100
	return Percentage(&usagePercentage)
}

func StorageUsagePercentageByFreeSpace(freeSpaceBytes *float64, storageSizeGB *int32) string {
	if freeSpaceBytes == nil || storageSizeGB == nil {
		return ""
	}
	storageSizeBytes := float64(*storageSizeGB) * (1024 * 1024 * 1024)
	usage := storageSizeBytes - *freeSpaceBytes
	usagePercentage := (usage / storageSizeBytes) * 100
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

func FormatFloat(number float64) string {
	isNegative := false
	if number < 0 {
		isNegative = true
		number = math.Abs(number)
	}
	parts := strings.Split(fmt.Sprintf("%.2f", number), ".")
	integerPart := parts[0]
	decimalPart := parts[1]

	var result []rune
	for i, digit := range integerPart {
		if i > 0 && (len(integerPart)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, rune(digit))
	}
	if isNegative {
		return fmt.Sprintf("-%s.%s", string(result), decimalPart)
	} else {
		return fmt.Sprintf("%s.%s", string(result), decimalPart)
	}
}
