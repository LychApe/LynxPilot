//go:build windows

package server

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	windowsKernel32          = windows.NewLazySystemDLL("kernel32.dll")
	procGetSystemTimes       = windowsKernel32.NewProc("GetSystemTimes")
	procGlobalMemoryStatusEx = windowsKernel32.NewProc("GlobalMemoryStatusEx")
	procGetTickCount64       = windowsKernel32.NewProc("GetTickCount64")
)

type windowsCPUTimes struct {
	total uint64
	idle  uint64
}

type windowsMemoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type windowsCurrentVersion struct {
	ProductName    string
	DisplayVersion string
	ReleaseID      string
	BuildNumber    string
	UBR            uint64
}

func collectPlatformStatus() Status {
	status := Status{}

	collectWindowsUptime(&status)
	collectWindowsLoad(&status.Load)
	appendWarning(&status, "读取 CPU 信息失败", collectCPU(&status.CPU))
	appendWarning(&status, "读取内存信息失败", collectMemory(&status.Memory))
	appendWarning(&status, "读取存储信息失败", collectStorage(&status.Storage))
	appendWarning(&status, "读取内核信息失败", collectKernel(&status.Kernel))
	appendWarning(&status, "读取系统版本信息失败", collectDistribution(&status.Distribution))

	return status
}

func collectWindowsUptime(status *Status) {
	r1, _, _ := procGetTickCount64.Call()
	milliseconds := uint64(r1)
	seconds := float64(milliseconds) / 1000.0
	status.UptimeSeconds = roundPercent(seconds)
	status.Uptime = formatUptime(seconds)
}

func collectWindowsLoad(load *LoadStatus) {
	// Windows does not expose Unix-style load averages. Keep these fields empty
	// and let the UI present them as not applicable.
	load.Load1 = 0
	load.Load5 = 0
	load.Load15 = 0
}

func collectCPU(cpu *CPUStatus) error {
	cpu.LogicalCores = runtime.NumCPU()
	collectWindowsCPUInfo(cpu)

	first, err := readWindowsCPUTimes()
	if err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)
	second, err := readWindowsCPUTimes()
	if err != nil {
		return err
	}

	total := second.total - first.total
	idle := second.idle - first.idle
	if total > 0 {
		cpu.UsagePercent = percentage(total-idle, total)
	}

	return nil
}

func collectWindowsCPUInfo(cpu *CPUStatus) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer key.Close()

	cpu.ModelName = registryString(key, "ProcessorNameString")
	cpu.VendorID = registryString(key, "VendorIdentifier")
}

func readWindowsCPUTimes() (windowsCPUTimes, error) {
	var idleTime windows.Filetime
	var kernelTime windows.Filetime
	var userTime windows.Filetime

	r1, _, callErr := procGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&idleTime)),
		uintptr(unsafe.Pointer(&kernelTime)),
		uintptr(unsafe.Pointer(&userTime)),
	)
	if r1 == 0 {
		return windowsCPUTimes{}, windowsCallError("GetSystemTimes", callErr)
	}

	return windowsCPUTimes{
		total: filetimeUint64(kernelTime) + filetimeUint64(userTime),
		idle:  filetimeUint64(idleTime),
	}, nil
}

func collectMemory(memory *MemoryStatus) error {
	status := windowsMemoryStatusEx{}
	status.Length = uint32(unsafe.Sizeof(status))

	r1, _, callErr := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&status)))
	if r1 == 0 {
		return windowsCallError("GlobalMemoryStatusEx", callErr)
	}

	memory.Total = status.TotalPhys
	memory.Available = status.AvailPhys
	memory.Free = status.AvailPhys
	if memory.Total > memory.Available {
		memory.Used = memory.Total - memory.Available
	}
	if status.TotalPageFile > status.TotalPhys {
		memory.SwapTotal = status.TotalPageFile - status.TotalPhys
	}
	if status.AvailPageFile > status.AvailPhys && memory.SwapTotal > 0 {
		availableSwap := status.AvailPageFile - status.AvailPhys
		if memory.SwapTotal > availableSwap {
			memory.SwapUsed = memory.SwapTotal - availableSwap
		}
	}
	memory.UsedPercent = percentage(memory.Used, memory.Total)
	memory.TotalText = formatBytes(memory.Total)
	memory.UsedText = formatBytes(memory.Used)
	memory.FreeText = formatBytes(memory.Free)
	memory.AvailableText = formatBytes(memory.Available)
	return nil
}

func collectStorage(storage *StorageStatus) error {
	drives, err := windowsLogicalDrives()
	if err != nil {
		return err
	}

	var lastErr error
	for _, drive := range drives {
		rootPath, err := windows.UTF16PtrFromString(drive)
		if err != nil {
			lastErr = err
			continue
		}
		if windows.GetDriveType(rootPath) != windows.DRIVE_FIXED {
			continue
		}

		var available uint64
		var total uint64
		var free uint64
		if err := windows.GetDiskFreeSpaceEx(rootPath, &available, &total, &free); err != nil {
			lastErr = err
			continue
		}

		used := total - free
		storage.Filesystems = append(storage.Filesystems, FilesystemStatus{
			Path:          drive,
			Device:        strings.TrimSuffix(drive, `\`),
			Filesystem:    windowsFilesystemName(rootPath),
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

	if len(storage.Filesystems) == 0 {
		if lastErr != nil {
			return lastErr
		}
		return errors.New("未找到固定磁盘")
	}

	return nil
}

func windowsLogicalDrives() ([]string, error) {
	buffer := make([]uint16, 256)
	n, err := windows.GetLogicalDriveStrings(uint32(len(buffer)), &buffer[0])
	if err != nil {
		return nil, err
	}
	if int(n) > len(buffer) {
		buffer = make([]uint16, n)
		n, err = windows.GetLogicalDriveStrings(uint32(len(buffer)), &buffer[0])
		if err != nil {
			return nil, err
		}
	}

	return splitWindowsMultiString(buffer[:n]), nil
}

func splitWindowsMultiString(buffer []uint16) []string {
	var values []string
	start := 0
	for i, char := range buffer {
		if char != 0 {
			continue
		}
		if i == start {
			break
		}
		values = append(values, windows.UTF16ToString(buffer[start:i]))
		start = i + 1
	}
	return values
}

func windowsFilesystemName(rootPath *uint16) string {
	var filesystem [256]uint16
	if err := windows.GetVolumeInformation(rootPath, nil, 0, nil, nil, nil, &filesystem[0], uint32(len(filesystem))); err != nil {
		return ""
	}
	return windows.UTF16ToString(filesystem[:])
}

func collectKernel(kernel *KernelStatus) error {
	version := windows.RtlGetVersion()
	kernel.OSType = "Windows"
	kernel.Release = fmt.Sprintf("%d.%d.%d", version.MajorVersion, version.MinorVersion, version.BuildNumber)
	kernel.Version = windowsProductName(version)
	return nil
}

func collectDistribution(distribution *DistributionStatus) error {
	version := windows.RtlGetVersion()
	currentVersion := readWindowsCurrentVersion()
	productName := normalizeWindowsProductName(currentVersion.ProductName, version)
	if productName == "" {
		productName = windowsProductName(version)
	}

	distribution.Name = productName
	distribution.ID = "windows"
	distribution.Version = firstNonEmpty(currentVersion.DisplayVersion, currentVersion.ReleaseID)
	distribution.VersionID = windowsBuildVersion(version, currentVersion)
	distribution.PrettyName = windowsPrettyName(productName, distribution.Version, distribution.VersionID)
	return nil
}

func readWindowsCurrentVersion() windowsCurrentVersion {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err != nil {
		key, err = registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
		if err != nil {
			return windowsCurrentVersion{}
		}
	}
	defer key.Close()

	return windowsCurrentVersion{
		ProductName:    registryString(key, "ProductName"),
		DisplayVersion: registryString(key, "DisplayVersion"),
		ReleaseID:      registryString(key, "ReleaseId"),
		BuildNumber:    registryString(key, "CurrentBuildNumber"),
		UBR:            registryInteger(key, "UBR"),
	}
}

func windowsProductName(version *windows.OsVersionInfoEx) string {
	if version.ProductType != 1 {
		return "Windows Server"
	}
	if version.MajorVersion == 10 && version.BuildNumber >= 22000 {
		return "Windows 11"
	}
	if version.MajorVersion == 10 {
		return "Windows 10"
	}
	if version.MajorVersion == 6 && version.MinorVersion == 3 {
		return "Windows 8.1"
	}
	if version.MajorVersion == 6 && version.MinorVersion == 2 {
		return "Windows 8"
	}
	if version.MajorVersion == 6 && version.MinorVersion == 1 {
		return "Windows 7"
	}
	return fmt.Sprintf("Windows NT %d.%d", version.MajorVersion, version.MinorVersion)
}

func normalizeWindowsProductName(productName string, version *windows.OsVersionInfoEx) string {
	productName = strings.TrimSpace(productName)
	if version.ProductType == 1 && version.MajorVersion == 10 && version.BuildNumber >= 22000 {
		return strings.Replace(productName, "Windows 10", "Windows 11", 1)
	}
	return productName
}

func windowsBuildVersion(version *windows.OsVersionInfoEx, currentVersion windowsCurrentVersion) string {
	build := firstNonEmpty(currentVersion.BuildNumber, strconv.FormatUint(uint64(version.BuildNumber), 10))
	if currentVersion.UBR > 0 {
		return build + "." + strconv.FormatUint(currentVersion.UBR, 10)
	}
	return build
}

func windowsPrettyName(productName string, version string, build string) string {
	parts := []string{productName}
	if version != "" {
		parts = append(parts, version)
	}
	prettyName := strings.Join(parts, " ")
	if build != "" {
		prettyName += " (Build " + build + ")"
	}
	return prettyName
}

func registryString(key registry.Key, name string) string {
	value, _, err := key.GetStringValue(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func registryInteger(key registry.Key, name string) uint64 {
	value, _, err := key.GetIntegerValue(name)
	if err != nil {
		return 0
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func filetimeUint64(filetime windows.Filetime) uint64 {
	return uint64(filetime.HighDateTime)<<32 | uint64(filetime.LowDateTime)
}

func windowsCallError(name string, err error) error {
	if err != nil && err != syscall.Errno(0) {
		return err
	}
	return errors.New(name + " 调用失败")
}
