package server

import (
	"os"
	"time"

	processService "github.com/LychApe/LynxPilot/internal/service/process"
	"github.com/LychApe/LynxPilot/internal/utils/logger"
	"github.com/LychApe/LynxPilot/internal/utils/response"
	"github.com/gin-gonic/gin"
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

func ShutdownHandler(c *gin.Context) {
	response.OK(c, gin.H{"message": "shutdown signal accepted"})

	go func() {
		triggerSelfShutdown()
	}()
}
