package setting

import (
	"fmt"

	settingModel "github.com/LychApe/LynxPilot/internal/model/setting"
	"gorm.io/gorm"
)

const (
	DockerHost       = "docker_host"
	DockerTLSVerify  = "docker_tls_verify"
	DockerCertPath   = "docker_cert_path"

	ContainerDefaultRestartPolicy = "container_default_restart_policy"
	ContainerDefaultLogDriver     = "container_default_log_driver"
	ContainerDefaultLogMaxSize    = "container_default_log_max_size"
	ContainerDefaultLogMaxFile    = "container_default_log_max_file"
	ContainerDefaultCPULimit      = "container_default_cpu_limit"
	ContainerDefaultMemoryLimit   = "container_default_memory_limit"
	ContainerAutoRefreshInterval  = "container_auto_refresh_interval"
	ContainerShowStoppedDefault   = "container_show_stopped_default"
)

type DockerConnection struct {
	Host      string `json:"host"`
	TLSVerify bool   `json:"tls_verify"`
	CertPath  string `json:"cert_path"`
}

type ContainerDefaults struct {
	RestartPolicy string `json:"restart_policy"`
	LogDriver     string `json:"log_driver"`
	LogMaxSize    string `json:"log_max_size"`
	LogMaxFile    int    `json:"log_max_file"`
	CPULimit      string `json:"cpu_limit"`
	MemoryLimit   string `json:"memory_limit"`
}

type ContainerUIPrefs struct {
	AutoRefreshInterval int  `json:"auto_refresh_interval"`
	ShowStoppedDefault  bool `json:"show_stopped_default"`
}

type AllSettings struct {
	Connection      DockerConnection  `json:"connection"`
	ContainerDefaults ContainerDefaults `json:"container_defaults"`
	UIPrefs         ContainerUIPrefs   `json:"ui_prefs"`
}

func GetDockerConnection(db *gorm.DB) (*DockerConnection, error) {
	values, err := settingModel.GetMulti(db, []string{DockerHost, DockerTLSVerify, DockerCertPath})
	if err != nil {
		return nil, err
	}

	conn := &DockerConnection{
		Host:     values[DockerHost],
		CertPath: values[DockerCertPath],
	}

	if values[DockerTLSVerify] == "true" {
		conn.TLSVerify = true
	}

	return conn, nil
}

func SaveDockerConnection(db *gorm.DB, conn *DockerConnection) error {
	tlsVal := "false"
	if conn.TLSVerify {
		tlsVal = "true"
	}

	return settingModel.SetMulti(db, map[string]string{
		DockerHost:      conn.Host,
		DockerTLSVerify: tlsVal,
		DockerCertPath:  conn.CertPath,
	})
}

func GetContainerDefaults(db *gorm.DB) (*ContainerDefaults, error) {
	keys := []string{
		ContainerDefaultRestartPolicy, ContainerDefaultLogDriver,
		ContainerDefaultLogMaxSize, ContainerDefaultLogMaxFile,
		ContainerDefaultCPULimit, ContainerDefaultMemoryLimit,
	}
	values, err := settingModel.GetMulti(db, keys)
	if err != nil {
		return nil, err
	}

	logMaxFile := 3
	if v, ok := values[ContainerDefaultLogMaxFile]; ok && v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			logMaxFile = n
		}
	}

	return &ContainerDefaults{
		RestartPolicy: values[ContainerDefaultRestartPolicy],
		LogDriver:     values[ContainerDefaultLogDriver],
		LogMaxSize:    values[ContainerDefaultLogMaxSize],
		LogMaxFile:    logMaxFile,
		CPULimit:      values[ContainerDefaultCPULimit],
		MemoryLimit:   values[ContainerDefaultMemoryLimit],
	}, nil
}

func SaveContainerDefaults(db *gorm.DB, defaults *ContainerDefaults) error {
	logMaxFile := "3"
	if defaults.LogMaxFile > 0 {
		logMaxFile = fmt.Sprintf("%d", defaults.LogMaxFile)
	}

	return settingModel.SetMulti(db, map[string]string{
		ContainerDefaultRestartPolicy: defaults.RestartPolicy,
		ContainerDefaultLogDriver:     defaults.LogDriver,
		ContainerDefaultLogMaxSize:    defaults.LogMaxSize,
		ContainerDefaultLogMaxFile:    logMaxFile,
		ContainerDefaultCPULimit:      defaults.CPULimit,
		ContainerDefaultMemoryLimit:   defaults.MemoryLimit,
	})
}

func GetContainerUIPrefs(db *gorm.DB) (*ContainerUIPrefs, error) {
	keys := []string{ContainerAutoRefreshInterval, ContainerShowStoppedDefault}
	values, err := settingModel.GetMulti(db, keys)
	if err != nil {
		return nil, err
	}

	interval := 10
	if v, ok := values[ContainerAutoRefreshInterval]; ok && v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n >= 0 {
			interval = n
		}
	}

	showStopped := true
	if v, ok := values[ContainerShowStoppedDefault]; ok && v == "false" {
		showStopped = false
	}

	return &ContainerUIPrefs{
		AutoRefreshInterval: interval,
		ShowStoppedDefault:  showStopped,
	}, nil
}

func SaveContainerUIPrefs(db *gorm.DB, prefs *ContainerUIPrefs) error {
	showStopped := "true"
	if !prefs.ShowStoppedDefault {
		showStopped = "false"
	}

	return settingModel.SetMulti(db, map[string]string{
		ContainerAutoRefreshInterval: fmt.Sprintf("%d", prefs.AutoRefreshInterval),
		ContainerShowStoppedDefault:  showStopped,
	})
}

func GetAllSettings(db *gorm.DB) (*AllSettings, error) {
	conn, err := GetDockerConnection(db)
	if err != nil {
		return nil, err
	}

	defaults, err := GetContainerDefaults(db)
	if err != nil {
		return nil, err
	}

	prefs, err := GetContainerUIPrefs(db)
	if err != nil {
		return nil, err
	}

	return &AllSettings{
		Connection:        *conn,
		ContainerDefaults: *defaults,
		UIPrefs:           *prefs,
	}, nil
}
