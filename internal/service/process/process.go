package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/LychApe/LynxPilot/internal/utils/logger"
)

func StartNewProcess() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	cmd := buildRestartCommand(executable)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return err
	}

	logger.Infof("已拉起新进程，PID=%d", cmd.Process.Pid)
	return nil
}

func buildRestartCommand(executable string) *exec.Cmd {
	// `go run` 启动时可执行文件在临时目录，优先回退到 go run 重启。
	if isGoRunTempBinary(executable) {
		repoRoot, err := findRepoRootFromCWD()
		if err == nil {
			cmd := exec.Command("go", "run", "./cmd/server/main.go")
			cmd.Dir = repoRoot
			return cmd
		}
		logger.Errorf("定位项目根目录失败，回退到原可执行文件重启: %v", err)
	}

	return exec.Command(executable, os.Args[1:]...)
}

func isGoRunTempBinary(executable string) bool {
	tempDir := filepath.Clean(os.TempDir())
	exe := filepath.Clean(executable)
	return strings.Contains(exe, string(filepath.Separator)+"go-build") &&
		strings.HasPrefix(exe, tempDir+string(filepath.Separator))
}

func findRepoRootFromCWD() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("未找到 go.mod")
		}
		dir = parent
	}
}
