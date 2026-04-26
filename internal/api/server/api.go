package server

import (
	"os"
	"runtime"
	"time"

	processService "github.com/LychApe/LynxPilot/internal/service/process"
	userService "github.com/LychApe/LynxPilot/internal/service/user"
	"github.com/LychApe/LynxPilot/internal/utils/logger"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/LychApe/LynxPilot/internal/utils/appvar"
	"github.com/LychApe/LynxPilot/internal/utils/format"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func triggerSelfShutdown() {
	pid := os.Getpid()
	process, err := os.FindProcess(pid)
	if err != nil {
		logger.Errorf("查找当前进程失败: %v", err)
		os.Exit(1)
	}

	if err := process.Signal(os.Interrupt); err != nil {
		logger.Errorf("发送关闭信号失败，执行兜底退出: %v", err)
		os.Exit(0)
	}
}

// 重启接口
func RebootHandler(c *gin.Context) {
	response.OK(c, gin.H{"message": "reboot signal accepted"})

	go func() {
		time.Sleep(200 * time.Millisecond)

		if err := processService.StartNewProcess(); err != nil {
			logger.Errorf("拉起新进程失败: %v", err)
			return
		}

		triggerSelfShutdown()
	}()
}

// 关闭接口
func ShutdownHandler(c *gin.Context) {
	response.OK(c, gin.H{"message": "shutdown signal accepted"})

	go func() {
		triggerSelfShutdown()
	}()
}

// 状态接口
func StatusHandler(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	uptime := time.Since(appvar.StartTime).Truncate(time.Second)

	response.OK(c, gin.H{
		"installed": userService.IsInstalled(db),
		"memory":    format.Memory(memStats.Alloc),
		"uptime":    uptime.String(),
	})
}
