//go:build linux

package server

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func collectPlatformStatus() Status {
	status := Status{}

	collectUptime(&status)
	appendWarning(&status, "读取运行负载失败", collectLoad(&status.Load))
	appendWarning(&status, "读取 CPU 信息失败", collectCPU(&status.CPU))
	appendWarning(&status, "读取内存信息失败", collectMemory(&status.Memory))
	appendWarning(&status, "读取存储信息失败", collectStorage(&status.Storage))
	appendWarning(&status, "读取内核信息失败", collectKernel(&status.Kernel))
	appendWarning(&status, "读取发行版信息失败", collectDistribution(&status.Distribution))

	return status
}

func collectUptime(status *Status) {
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return
	}

	fields := strings.Fields(string(content))
	if len(fields) < 1 {
		return
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return
	}

	status.UptimeSeconds = roundPercent(seconds)
	status.Uptime = formatUptime(seconds)
}

func collectLoad(load *LoadStatus) error {
	content, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return err
	}

	fields := strings.Fields(string(content))
	if len(fields) < 5 {
		return errors.New("/proc/loadavg 格式无效")
	}

	load1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return err
	}
	load5, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return err
	}
	load15, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return err
	}

	processes := strings.Split(fields[3], "/")
	if len(processes) == 2 {
		load.RunningProcesses, _ = strconv.Atoi(processes[0])
		load.TotalProcesses, _ = strconv.Atoi(processes[1])
	}

	load.LastPID, _ = strconv.Atoi(fields[4])
	load.Load1 = roundPercent(load1)
	load.Load5 = roundPercent(load5)
	load.Load15 = roundPercent(load15)
	return nil
}

func collectCPU(cpu *CPUStatus) error {
	info, infoErr := parseProcCPUInfo()
	if modelName := firstValue(info, "model name", "Hardware", "Processor"); modelName != "" {
		cpu.ModelName = modelName
	}
	cpu.VendorID = firstValue(info, "vendor_id")
	cpu.PhysicalCores = physicalCoreCount(info)

	first, err := readCPUTimes()
	if err != nil {
		if infoErr != nil {
			return errors.Join(infoErr, err)
		}
		return err
	}
	time.Sleep(200 * time.Millisecond)
	second, err := readCPUTimes()
	if err != nil {
		if infoErr != nil {
			return errors.Join(infoErr, err)
		}
		return err
	}

	total := second.total - first.total
	idle := second.idle - first.idle
	if total > 0 {
		cpu.UsagePercent = percentage(total-idle, total)
	}

	return infoErr
}

func parseProcCPUInfo() (map[string][]string, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info := make(map[string][]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			continue
		}
		info[key] = append(info[key], value)
	}
	if err := scanner.Err(); err != nil {
		return info, err
	}
	return info, nil
}

func firstValue(info map[string][]string, keys ...string) string {
	for _, key := range keys {
		values := info[key]
		if len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func physicalCoreCount(info map[string][]string) int {
	physicalIDs := info["physical id"]
	coreIDs := info["core id"]
	if len(physicalIDs) > 0 && len(physicalIDs) == len(coreIDs) {
		cores := make(map[string]struct{}, len(physicalIDs))
		for i := range physicalIDs {
			cores[physicalIDs[i]+":"+coreIDs[i]] = struct{}{}
		}
		return len(cores)
	}

	cpuCores := info["cpu cores"]
	if len(cpuCores) == 0 {
		return 0
	}
	cores, _ := strconv.Atoi(cpuCores[0])
	return cores
}

type cpuTimes struct {
	total uint64
	idle  uint64
}

func readCPUTimes() (cpuTimes, error) {
	content, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuTimes{}, err
	}

	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != "cpu" {
			continue
		}

		var total uint64
		var values []uint64
		for _, field := range fields[1:] {
			value, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				return cpuTimes{}, err
			}
			values = append(values, value)
			total += value
		}

		var idle uint64
		if len(values) > 3 {
			idle = values[3]
		}
		if len(values) > 4 {
			idle += values[4]
		}
		return cpuTimes{total: total, idle: idle}, nil
	}

	return cpuTimes{}, errors.New("/proc/stat 缺少 cpu 汇总行")
}

func collectMemory(memory *MemoryStatus) error {
	content, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return err
	}

	values := make(map[string]uint64)
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		values[key] = value * 1024
	}

	memory.Total = values["MemTotal"]
	memory.Free = values["MemFree"]
	memory.Available = values["MemAvailable"]
	memory.Buffers = values["Buffers"]
	memory.Cached = values["Cached"] + values["SReclaimable"]
	memory.SwapTotal = values["SwapTotal"]
	if memory.SwapTotal > values["SwapFree"] {
		memory.SwapUsed = memory.SwapTotal - values["SwapFree"]
	}
	if memory.Total > memory.Available {
		memory.Used = memory.Total - memory.Available
	}
	memory.UsedPercent = percentage(memory.Used, memory.Total)
	memory.TotalText = formatBytes(memory.Total)
	memory.UsedText = formatBytes(memory.Used)
	memory.FreeText = formatBytes(memory.Free)
	memory.AvailableText = formatBytes(memory.Available)
	return nil
}

func collectStorage(storage *StorageStatus) error {
	mounts, err := readMounts()
	if err != nil {
		return err
	}

	seen := map[string]struct{}{}
	for _, mount := range mounts {
		if _, ok := seen[mount.Path]; ok {
			continue
		}
		seen[mount.Path] = struct{}{}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mount.Path, &stat); err != nil {
			continue
		}

		total := stat.Blocks * uint64(stat.Bsize)
		free := stat.Bfree * uint64(stat.Bsize)
		available := stat.Bavail * uint64(stat.Bsize)
		used := total - free

		storage.Filesystems = append(storage.Filesystems, FilesystemStatus{
			Path:          mount.Path,
			Device:        mount.Device,
			Filesystem:    mount.Filesystem,
			Total:         total,
			Used:          used,
			Free:          free,
			Available:     available,
			UsedPercent:   percentage(used, total),
			TotalText:     formatBytes(total),
			UsedText:      formatBytes(used),
			FreeText:      formatBytes(free),
			AvailableText: formatBytes(available),
		})
	}

	return nil
}

type mountInfo struct {
	Device     string
	Path       string
	Filesystem string
}

func readMounts() ([]mountInfo, error) {
	content, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil, err
	}

	virtualFilesystems := map[string]struct{}{
		"autofs":      {},
		"binfmt_misc": {},
		"bpf":         {},
		"cgroup":      {},
		"cgroup2":     {},
		"configfs":    {},
		"debugfs":     {},
		"devpts":      {},
		"devtmpfs":    {},
		"fusectl":     {},
		"hugetlbfs":   {},
		"mqueue":      {},
		"proc":        {},
		"pstore":      {},
		"rpc_pipefs":  {},
		"securityfs":  {},
		"sysfs":       {},
		"tmpfs":       {},
		"tracefs":     {},
	}

	var mounts []mountInfo
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		if _, ok := virtualFilesystems[fields[2]]; ok {
			continue
		}
		mounts = append(mounts, mountInfo{
			Device:     unescapeMountField(fields[0]),
			Path:       unescapeMountField(fields[1]),
			Filesystem: fields[2],
		})
	}

	return mounts, nil
}

func unescapeMountField(value string) string {
	replacer := strings.NewReplacer(`\040`, " ", `\011`, "\t", `\012`, "\n", `\134`, `\`)
	return replacer.Replace(value)
}

func collectKernel(kernel *KernelStatus) error {
	var uts syscall.Utsname
	if err := syscall.Uname(&uts); err != nil {
		return err
	}

	kernel.OSType = utsString(uts.Sysname[:])
	kernel.Release = utsString(uts.Release[:])
	kernel.Version = utsString(uts.Version[:])
	return nil
}

func utsString(chars []int8) string {
	var builder strings.Builder
	for _, char := range chars {
		if char == 0 {
			break
		}
		builder.WriteByte(byte(char))
	}
	return builder.String()
}

func collectDistribution(distribution *DistributionStatus) error {
	values, err := readOSRelease()
	if err != nil {
		return err
	}

	distribution.Name = values["NAME"]
	distribution.ID = values["ID"]
	distribution.Version = values["VERSION"]
	distribution.VersionID = values["VERSION_ID"]
	distribution.PrettyName = values["PRETTY_NAME"]
	return nil
}

func readOSRelease() (map[string]string, error) {
	paths := []string{"/etc/os-release", "/usr/lib/os-release"}
	var lastErr error
	for _, path := range paths {
		values, err := parseOSRelease(path)
		if err == nil {
			return values, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func parseOSRelease(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		parsedValue, err := strconv.Unquote(value)
		if err != nil {
			parsedValue = value
		}
		values[key] = parsedValue
	}
	if err := scanner.Err(); err != nil {
		return values, err
	}
	return values, nil
}
