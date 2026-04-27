//go:build !linux && !windows

package server

func collectPlatformStatus() Status {
	return Status{
		Warnings: []string{"当前平台暂未支持主机状态采集"},
	}
}
