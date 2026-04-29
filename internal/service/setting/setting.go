package setting

import (
	settingModel "github.com/LychApe/LynxPilot/internal/model/setting"
	"gorm.io/gorm"
)

const (
	DockerHost       = "docker_host"
	DockerTLSVerify  = "docker_tls_verify"
	DockerCertPath   = "docker_cert_path"
)

type DockerConnection struct {
	Host      string `json:"host"`
	TLSVerify bool   `json:"tls_verify"`
	CertPath  string `json:"cert_path"`
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
