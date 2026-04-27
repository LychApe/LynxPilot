package server

import (
	"math"
	"net"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

type Status struct {
	IPAddresses  []string           `json:"ip_addresses"`
	Uptime       string             `json:"uptime"`
	UptimeSeconds float64           `json:"uptime_seconds"`
	Load         LoadStatus         `json:"load"`
	CPU          CPUStatus          `json:"cpu"`
	Memory       MemoryStatus       `json:"memory"`
	Storage      StorageStatus      `json:"storage"`
	Kernel       KernelStatus       `json:"kernel"`
	Distribution DistributionStatus `json:"distribution"`
	Warnings     []string           `json:"warnings,omitempty"`
}

type LoadStatus struct {
	Load1            float64 `json:"load1"`
	Load5            float64 `json:"load5"`
	Load15           float64 `json:"load15"`
	RunningProcesses int     `json:"running_processes,omitempty"`
	TotalProcesses   int     `json:"total_processes,omitempty"`
	LastPID          int     `json:"last_pid,omitempty"`
	PerCoreLoad1     float64 `json:"per_core_load1"`
}

type CPUStatus struct {
	UsagePercent  float64 `json:"usage_percent"`
	LogicalCores  int     `json:"logical_cores"`
	PhysicalCores int     `json:"physical_cores,omitempty"`
	ModelName     string  `json:"model_name,omitempty"`
	VendorID      string  `json:"vendor_id,omitempty"`
	Architecture  string  `json:"architecture"`
}

type MemoryStatus struct {
	Total         uint64  `json:"total"`
	Used          uint64  `json:"used"`
	Free          uint64  `json:"free"`
	Available     uint64  `json:"available"`
	Buffers       uint64  `json:"buffers,omitempty"`
	Cached        uint64  `json:"cached,omitempty"`
	SwapTotal     uint64  `json:"swap_total,omitempty"`
	SwapUsed      uint64  `json:"swap_used,omitempty"`
	UsedPercent   float64 `json:"used_percent"`
	TotalText     string  `json:"total_text"`
	UsedText      string  `json:"used_text"`
	FreeText      string  `json:"free_text"`
	AvailableText string  `json:"available_text"`
}

type StorageStatus struct {
	Filesystems []FilesystemStatus `json:"filesystems"`
}

type FilesystemStatus struct {
	Path          string  `json:"path"`
	Device        string  `json:"device,omitempty"`
	Filesystem    string  `json:"filesystem,omitempty"`
	Total         uint64  `json:"total"`
	Used          uint64  `json:"used"`
	Free          uint64  `json:"free"`
	Available     uint64  `json:"available"`
	UsedPercent   float64 `json:"used_percent"`
	TotalText     string  `json:"total_text"`
	UsedText      string  `json:"used_text"`
	FreeText      string  `json:"free_text"`
	AvailableText string  `json:"available_text"`
}

type KernelStatus struct {
	OSType  string `json:"os_type,omitempty"`
	Release string `json:"release,omitempty"`
	Version string `json:"version,omitempty"`
	GOOS    string `json:"goos"`
	GOARCH  string `json:"goarch"`
}

type DistributionStatus struct {
	Name       string `json:"name,omitempty"`
	ID         string `json:"id,omitempty"`
	Version    string `json:"version,omitempty"`
	VersionID  string `json:"version_id,omitempty"`
	PrettyName string `json:"pretty_name,omitempty"`
}

func GetStatus() Status {
	status := collectPlatformStatus()

	status.CPU.LogicalCores = runtime.NumCPU()
	status.CPU.Architecture = runtime.GOARCH
	status.Kernel.GOOS = runtime.GOOS
	status.Kernel.GOARCH = runtime.GOARCH
	status.Load.PerCoreLoad1 = roundPercent(status.Load.Load1 / float64(max(status.CPU.LogicalCores, 1)))

	status.IPAddresses = collectIPAddresses()

	return status
}

func appendWarning(status *Status, message string, err error) {
	if err == nil {
		return
	}
	status.Warnings = append(status.Warnings, message+": "+err.Error())
}

func percentage(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return roundPercent(float64(used) * 100 / float64(total))
}

func roundPercent(value float64) float64 {
	return math.Round(value*100) / 100
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return formatFloat(float64(bytes), "B")
	}

	value := float64(bytes)
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	for _, suffix := range units {
		value /= unit
		if value < unit {
			return formatFloat(value, suffix)
		}
	}

	return formatFloat(value, "EB")
}

func formatFloat(value float64, suffix string) string {
	value = math.Round(value*10) / 10
	if value == math.Trunc(value) {
		return strconv.FormatFloat(value, 'f', 0, 64) + suffix
	}
	return strconv.FormatFloat(value, 'f', 1, 64) + suffix
}

func collectIPAddresses() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	var addresses []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ip := addr.String()
			idx := strings.Index(ip, "/")
			if idx >= 0 {
				ip = ip[:idx]
			}
			if strings.Contains(ip, ":") {
				continue
			}
			if ip != "" && ip != "0.0.0.0" {
				addresses = append(addresses, ip)
			}
		}
	}

	sort.Strings(addresses)
	return addresses
}

func formatUptime(seconds float64) string {
	totalSec := int64(seconds)
	days := totalSec / 86400
	hours := (totalSec % 86400) / 3600
	mins := (totalSec % 3600) / 60
	secs := totalSec % 60

	var parts []string
	if days > 0 {
		parts = append(parts, strconv.FormatInt(days, 10)+" 天")
	}
	if hours > 0 {
		parts = append(parts, strconv.FormatInt(hours, 10)+" 小时")
	}
	if mins > 0 {
		parts = append(parts, strconv.FormatInt(mins, 10)+" 分钟")
	}
	if secs > 0 || len(parts) == 0 {
		parts = append(parts, strconv.FormatInt(secs, 10)+" 秒")
	}
	return strings.Join(parts, " ")
}
