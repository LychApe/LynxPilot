package bootstrap

import (
	"os"
)

func Run() {
	// 加载配置文件
	config, err := LoadConfig("")
	if err != nil {
		os.Exit(1)
	}

	// 初始化数据库
	if _, err := LoadDatabase(config); err != nil {
		os.Exit(1)
	}

	// 加载路由
	LoadRouter(config)
}
