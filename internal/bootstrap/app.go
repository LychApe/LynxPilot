package bootstrap

import (
	"os"
)

func Run() {
	// 加载配置文件
	Config, err := LoadConfig("")
	if err != nil {
		os.Exit(1)
	}
	// 加载路由
	LoadRouter(Config)
}
