package docker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"gorm.io/gorm"
)

type ConnectionConfig struct {
	Host      string
	TLSVerify bool
	CertPath  string
}

var (
	configMu sync.RWMutex
	activeDB *gorm.DB
)

func SetDB(db *gorm.DB) {
	configMu.Lock()
	defer configMu.Unlock()
	activeDB = db
}

func getDB() *gorm.DB {
	configMu.RLock()
	defer configMu.RUnlock()
	return activeDB
}

func loadConnectionConfig() *ConnectionConfig {
	db := getDB()
	if db == nil {
		return nil
	}

	var settings []struct {
		Key   string `gorm:"column:key"`
		Value string `gorm:"column:value"`
	}
	db.Table("settings").Where("`key` IN ?", []string{"docker_host", "docker_tls_verify", "docker_cert_path"}).Find(&settings)

	m := make(map[string]string, len(settings))
	for _, s := range settings {
		m[s.Key] = s.Value
	}

	if m["docker_host"] == "" {
		return nil
	}

	cfg := &ConnectionConfig{
		Host:      m["docker_host"],
		TLSVerify: m["docker_tls_verify"] == "true",
		CertPath:  m["docker_cert_path"],
	}
	return cfg
}

func buildClient(cfg *ConnectionConfig) (*client.Client, error) {
	if cfg == nil || cfg.Host == "" {
		return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	}

	hostURL := cfg.Host

	opts := []client.Opt{client.WithAPIVersionNegotiation(), client.WithHost(hostURL)}

	useTLS := cfg.TLSVerify

	if !useTLS && isTLSPort(hostURL) {
		useTLS = true
	}

	if useTLS {
		httpClient, err := buildTLSHttpClient(cfg.CertPath)
		if err != nil {
			return nil, fmt.Errorf("TLS 配置失败: %w", err)
		}
		opts = append(opts, client.WithHTTPClient(httpClient))

		if strings.HasPrefix(hostURL, "tcp://") {
			opts[1] = client.WithHost("https://" + strings.TrimPrefix(hostURL, "tcp://"))
		}
	}

	return client.NewClientWithOpts(opts...)
}

func isTLSPort(host string) bool {
	u, err := url.Parse(host)
	if err != nil {
		return false
	}
	_, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return false
	}
	return portStr == "2376"
}

func buildTLSHttpClient(certPath string) (*http.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: certPath == "",
	}

	if certPath != "" {
		caPath := filepath.Join(certPath, "ca.pem")
		certFile := filepath.Join(certPath, "cert.pem")
		keyFile := filepath.Join(certPath, "key.pem")

		for _, f := range []string{caPath, certFile, keyFile} {
			if _, err := os.Stat(f); os.IsNotExist(err) {
				return nil, fmt.Errorf("证书文件不存在: %s", f)
			}
		}

		caCert, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("读取 CA 证书失败: %w", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("加载客户端证书失败: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.InsecureSkipVerify = false
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}, nil
}

func getClient() (*client.Client, error) {
	cfg := loadConnectionConfig()
	cli, err := buildClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("连接 Docker 守护进程失败: %w", err)
	}
	return cli, nil
}

func TestConnection(cfg *ConnectionConfig) error {
	if cfg == nil || cfg.Host == "" {
		cfg = &ConnectionConfig{}
	}

	cli, err := buildClient(cfg)
	if err != nil {
		return err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = cli.Ping(ctx)
	if err != nil {
		return enrichConnectionError(cfg, err)
	}
	return nil
}

func enrichConnectionError(cfg *ConnectionConfig, err error) error {
	msg := err.Error()

	if strings.Contains(msg, "certificate") || strings.Contains(msg, "tls") || strings.Contains(msg, "x509") {
		return fmt.Errorf("%s\n提示: 端口 2376 需要 TLS 证书，请在连接设置中启用 TLS 并配置证书目录", msg)
	}

	if strings.Contains(msg, "connection refused") {
		return fmt.Errorf("%s\n提示: 请确认远程 Docker 已开启 TCP 监听 (端口 %s)", msg, extractPort(cfg.Host))
	}

	if strings.Contains(msg, "i/o timeout") || strings.Contains(msg, "deadline exceeded") || strings.Contains(msg, "context deadline") {
		return fmt.Errorf("%s\n提示: 连接超时，请检查网络是否可达 %s", msg, cfg.Host)
	}

	if strings.Contains(msg, "no such file") || strings.Contains(msg, "cannot find the file") {
		return fmt.Errorf("%s\n提示: 请确认 Docker Socket 文件路径正确", msg)
	}

	return fmt.Errorf("连接 Docker 失败: %s (Host: %s)", msg, cfg.Host)
}

func extractPort(host string) string {
	u, err := url.Parse(host)
	if err != nil {
		return "unknown"
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return "unknown"
	}
	return port
}

func GetActiveHost() string {
	cfg := loadConnectionConfig()
	if cfg != nil && cfg.Host != "" {
		return cfg.Host
	}
	return "env://" + os.Getenv("DOCKER_HOST")
}

func IsCustomConnection() bool {
	cfg := loadConnectionConfig()
	return cfg != nil && cfg.Host != ""
}

type ContainerInfo struct {
	ID      string   `json:"id"`
	Names   []string `json:"names"`
	Image   string   `json:"image"`
	State   string   `json:"state"`
	Status  string   `json:"status"`
	Created int64    `json:"created"`
	Ports   []Port   `json:"ports"`
	Command string   `json:"command"`
}

type Port struct {
	IP          string `json:"ip"`
	PrivatePort uint16 `json:"private_port"`
	PublicPort  uint16 `json:"public_port"`
	Type        string `json:"type"`
}

type ContainerStats struct {
	CPUPercent      float64 `json:"cpu_percent"`
	MemoryUsage     uint64  `json:"memory_usage"`
	MemoryLimit     uint64  `json:"memory_limit"`
	MemoryPercent   float64 `json:"memory_percent"`
	NetworkRx       uint64  `json:"network_rx"`
	NetworkTx       uint64  `json:"network_tx"`
	BlockRead       uint64  `json:"block_read"`
	BlockWrite      uint64  `json:"block_write"`
	Pids            int     `json:"pids"`
	MemoryUsageText string  `json:"memory_usage_text"`
	MemoryLimitText string  `json:"memory_limit_text"`
}

type ImageInfo struct {
	ID          string            `json:"id"`
	ShortID     string            `json:"short_id"`
	RepoTags    []string          `json:"repo_tags"`
	RepoDigests []string          `json:"repo_digests"`
	Size        int64             `json:"size"`
	SizeText    string            `json:"size_text"`
	Created     int64             `json:"created"`
	Containers  int64             `json:"containers"`
	Labels      map[string]string `json:"labels"`
}

type PullImageRequest struct {
	Image string `json:"image"`
}

type TagImageRequest struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type PruneResult struct {
	Deleted        int    `json:"deleted"`
	SpaceReclaimed uint64 `json:"space_reclaimed"`
	SpaceText      string `json:"space_text"`
}

type RegistryConfig struct {
	Name          string `json:"name"`
	ServerAddress string `json:"server_address"`
	Username      string `json:"username"`
	Password      string `json:"password,omitempty"`
}

type VolumeInfo struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	CreatedAt  string            `json:"created_at"`
	Scope      string            `json:"scope"`
	Labels     map[string]string `json:"labels"`
	Options    map[string]string `json:"options"`
	Size       int64             `json:"size"`
	SizeText   string            `json:"size_text"`
}

type CreateVolumeRequest struct {
	Name    string            `json:"name"`
	Driver  string            `json:"driver"`
	Labels  map[string]string `json:"labels"`
	Options map[string]string `json:"options"`
}

type ContainerDetail struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Image        string   `json:"image"`
	State        string   `json:"state"`
	Status       string   `json:"status"`
	Created      string   `json:"created"`
	Command      string   `json:"command"`
	Env          []string `json:"env"`
	Ports        []Port   `json:"ports"`
	NetworkMode  string   `json:"network_mode"`
	IPAddress    string   `json:"ip_address"`
	Gateway      string   `json:"gateway"`
	MacAddress   string   `json:"mac_address"`
	RestartCount int      `json:"restart_count"`
	StartedAt    string   `json:"started_at"`
	FinishedAt   string   `json:"finished_at"`
}

func ListContainers(all bool) ([]ContainerInfo, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: all})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %w", err)
	}

	result := make([]ContainerInfo, 0, len(containers))
	for _, c := range containers {
		info := ContainerInfo{
			ID:      c.ID[:12],
			Names:   c.Names,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: c.Created,
			Command: c.Command,
		}

		for _, p := range c.Ports {
			info.Ports = append(info.Ports, Port{
				IP:          p.IP,
				PrivatePort: p.PrivatePort,
				PublicPort:  p.PublicPort,
				Type:        p.Type,
			})
		}

		result = append(result, info)
	}

	return result, nil
}

func GetContainerDetail(containerID string) (*ContainerDetail, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	inspect, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return nil, fmt.Errorf("获取容器详情失败: %w", err)
	}

	detail := &ContainerDetail{
		ID:           inspect.ID[:12],
		Name:         strings.TrimPrefix(inspect.Name, "/"),
		Image:        inspect.Config.Image,
		State:        string(inspect.State.Status),
		Status:       string(inspect.State.Status),
		Created:      inspect.Created,
		Command:      strings.Join(inspect.Config.Cmd, " "),
		Env:          inspect.Config.Env,
		NetworkMode:  string(inspect.HostConfig.NetworkMode),
		RestartCount: inspect.RestartCount,
		StartedAt:    inspect.State.StartedAt,
		FinishedAt:   inspect.State.FinishedAt,
	}

	if inspect.NetworkSettings != nil {
		for name, net := range inspect.NetworkSettings.Networks {
			if name == "bridge" || name == "host" || len(inspect.NetworkSettings.Networks) == 1 {
				detail.IPAddress = net.IPAddress
				detail.Gateway = net.Gateway
				detail.MacAddress = net.MacAddress
				break
			}
		}

		for _, b := range inspect.NetworkSettings.Ports {
			for _, binding := range b {
				pPort, _ := strconv.ParseUint(binding.HostPort, 10, 16)
				detail.Ports = append(detail.Ports, Port{
					IP:         binding.HostIP,
					PublicPort: uint16(pPort),
					Type:       "tcp",
				})
			}
		}
	}

	if inspect.State.Running {
		startedAt, parseErr := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
		if parseErr == nil {
			uptime := time.Since(startedAt)
			detail.Status = fmt.Sprintf("运行中 %s", formatDuration(uptime))
		}
	}

	return detail, nil
}

func GetContainerStats(containerID string) (*ContainerStats, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	stream, err := cli.ContainerStats(context.Background(), containerID, false)
	if err != nil {
		return nil, fmt.Errorf("获取容器状态失败: %w", err)
	}
	defer stream.Body.Close()

	body, err := io.ReadAll(stream.Body)
	if err != nil {
		return nil, fmt.Errorf("读取容器状态数据失败: %w", err)
	}

	var stats container.StatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("解析容器状态数据失败: %w", err)
	}

	cpuPercent := calculateCPUPercent(&stats)
	memPercent := float64(0)
	if stats.MemoryStats.Limit > 0 {
		memPercent = roundPercent(float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100)
	}

	netRx, netTx := calculateNetworkStats(&stats)
	blockRead, blockWrite := calculateBlockStats(&stats)

	return &ContainerStats{
		CPUPercent:      cpuPercent,
		MemoryUsage:     stats.MemoryStats.Usage,
		MemoryLimit:     stats.MemoryStats.Limit,
		MemoryPercent:   memPercent,
		NetworkRx:       netRx,
		NetworkTx:       netTx,
		BlockRead:       blockRead,
		BlockWrite:      blockWrite,
		Pids:            int(stats.PidsStats.Current),
		MemoryUsageText: formatBytes(stats.MemoryStats.Usage),
		MemoryLimitText: formatBytes(stats.MemoryStats.Limit),
	}, nil
}

func StartContainer(containerID string) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	return cli.ContainerStart(context.Background(), containerID, container.StartOptions{})
}

func StopContainer(containerID string) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	timeout := 10
	return cli.ContainerStop(context.Background(), containerID, container.StopOptions{Timeout: &timeout})
}

func RestartContainer(containerID string) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	timeout := 10
	return cli.ContainerRestart(context.Background(), containerID, container.StopOptions{Timeout: &timeout})
}

func RemoveContainer(containerID string, force bool) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	return cli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: force})
}

func GetContainerLogs(containerID string, tail string) (string, error) {
	cli, err := getClient()
	if err != nil {
		return "", err
	}
	defer cli.Close()

	if tail == "" {
		tail = "100"
	}

	out, err := cli.ContainerLogs(context.Background(), containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
		Timestamps: true,
	})
	if err != nil {
		return "", fmt.Errorf("获取容器日志失败: %w", err)
	}
	defer out.Close()

	body, err := io.ReadAll(out)
	if err != nil {
		return "", fmt.Errorf("读取日志数据失败: %w", err)
	}

	return string(body), nil
}

func ListImages() ([]ImageInfo, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	images, err := cli.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取镜像列表失败: %w", err)
	}

	result := make([]ImageInfo, 0, len(images))
	for _, img := range images {
		id := strings.TrimPrefix(img.ID, "sha256:")
		shortID := id
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		result = append(result, ImageInfo{
			ID:          img.ID,
			ShortID:     shortID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Size:        img.Size,
			SizeText:    formatBytes(uint64(max(img.Size, 0))),
			Created:     img.Created,
			Containers:  img.Containers,
			Labels:      img.Labels,
		})
	}

	return result, nil
}

func PullImage(imageName string, registryConfig *RegistryConfig) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	options := image.PullOptions{}
	if registryConfig != nil && registryConfig.Username != "" {
		auth, err := registry.EncodeAuthConfig(registry.AuthConfig{
			Username:      registryConfig.Username,
			Password:      registryConfig.Password,
			ServerAddress: registryConfig.ServerAddress,
		})
		if err != nil {
			return fmt.Errorf("生成仓库认证失败: %w", err)
		}
		options.RegistryAuth = auth
	}

	reader, err := cli.ImagePull(context.Background(), imageName, options)
	if err != nil {
		return fmt.Errorf("拉取镜像失败: %w", err)
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("读取拉取结果失败: %w", err)
	}
	return nil
}

func RemoveImage(imageID string, force bool) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	_, err = cli.ImageRemove(context.Background(), imageID, image.RemoveOptions{Force: force, PruneChildren: true})
	if err != nil {
		return fmt.Errorf("删除镜像失败: %w", err)
	}
	return nil
}

func TagImage(source string, target string) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := cli.ImageTag(context.Background(), source, target); err != nil {
		return fmt.Errorf("镜像打标签失败: %w", err)
	}
	return nil
}

func PruneImages() (*PruneResult, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	report, err := cli.ImagesPrune(context.Background(), filters.NewArgs())
	if err != nil {
		return nil, fmt.Errorf("清理镜像失败: %w", err)
	}

	return &PruneResult{
		Deleted:        len(report.ImagesDeleted),
		SpaceReclaimed: report.SpaceReclaimed,
		SpaceText:      formatBytes(report.SpaceReclaimed),
	}, nil
}

func TestRegistry(config RegistryConfig) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	_, err = cli.RegistryLogin(context.Background(), registry.AuthConfig{
		Username:      config.Username,
		Password:      config.Password,
		ServerAddress: config.ServerAddress,
	})
	if err != nil {
		return fmt.Errorf("仓库登录失败: %w", err)
	}
	return nil
}

func ListVolumes() ([]VolumeInfo, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	res, err := cli.VolumeList(context.Background(), volume.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取存储卷列表失败: %w", err)
	}

	volumes := make([]VolumeInfo, 0, len(res.Volumes))
	for _, v := range res.Volumes {
		if v == nil {
			continue
		}
		size := int64(-1)
		sizeText := "-"
		if v.UsageData != nil {
			size = v.UsageData.Size
			if size >= 0 {
				sizeText = formatBytes(uint64(size))
			}
		}
		volumes = append(volumes, VolumeInfo{
			Name:       v.Name,
			Driver:     v.Driver,
			Mountpoint: v.Mountpoint,
			CreatedAt:  v.CreatedAt,
			Scope:      v.Scope,
			Labels:     v.Labels,
			Options:    v.Options,
			Size:       size,
			SizeText:   sizeText,
		})
	}
	return volumes, nil
}

func CreateVolume(req *CreateVolumeRequest) (*VolumeInfo, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	driver := strings.TrimSpace(req.Driver)
	if driver == "" {
		driver = "local"
	}
	vol, err := cli.VolumeCreate(context.Background(), volume.CreateOptions{
		Name:       req.Name,
		Driver:     driver,
		Labels:     req.Labels,
		DriverOpts: req.Options,
	})
	if err != nil {
		return nil, fmt.Errorf("创建存储卷失败: %w", err)
	}

	return &VolumeInfo{
		Name:       vol.Name,
		Driver:     vol.Driver,
		Mountpoint: vol.Mountpoint,
		CreatedAt:  vol.CreatedAt,
		Scope:      vol.Scope,
		Labels:     vol.Labels,
		Options:    vol.Options,
		Size:       -1,
		SizeText:   "-",
	}, nil
}

func RemoveVolume(name string, force bool) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	if err := cli.VolumeRemove(context.Background(), name, force); err != nil {
		return fmt.Errorf("删除存储卷失败: %w", err)
	}
	return nil
}

func PruneVolumes() (*PruneResult, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	report, err := cli.VolumesPrune(context.Background(), filters.NewArgs())
	if err != nil {
		return nil, fmt.Errorf("清理存储卷失败: %w", err)
	}
	return &PruneResult{
		Deleted:        len(report.VolumesDeleted),
		SpaceReclaimed: report.SpaceReclaimed,
		SpaceText:      formatBytes(report.SpaceReclaimed),
	}, nil
}

func CheckDockerAvailable() error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	_, err = cli.Ping(context.Background())
	return err
}

func SearchContainersByName(name string) ([]ContainerInfo, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	args := filters.NewArgs()
	args.Add("name", name)

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return nil, fmt.Errorf("搜索容器失败: %w", err)
	}

	result := make([]ContainerInfo, 0, len(containers))
	for _, c := range containers {
		info := ContainerInfo{
			ID:      c.ID[:12],
			Names:   c.Names,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: c.Created,
			Command: c.Command,
		}
		for _, p := range c.Ports {
			info.Ports = append(info.Ports, Port{
				IP:          p.IP,
				PrivatePort: p.PrivatePort,
				PublicPort:  p.PublicPort,
				Type:        p.Type,
			})
		}
		result = append(result, info)
	}

	return result, nil
}

func calculateCPUPercent(stats *container.StatsResponse) float64 {
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0 && cpuDelta > 0 {
		cpuPercent := (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100
		return roundPercent(cpuPercent)
	}
	return 0
}

func calculateNetworkStats(stats *container.StatsResponse) (rx uint64, tx uint64) {
	for _, network := range stats.Networks {
		rx += network.RxBytes
		tx += network.TxBytes
	}
	return
}

func calculateBlockStats(stats *container.StatsResponse) (read uint64, write uint64) {
	for _, bio := range stats.BlkioStats.IoServiceBytesRecursive {
		switch bio.Op {
		case "read":
			read += bio.Value
		case "write":
			write += bio.Value
		}
	}
	return
}

func roundPercent(value float64) float64 {
	return math.Round(value*100) / 100
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	value := float64(bytes)
	units := []string{"KB", "MB", "GB", "TB"}
	for _, suffix := range units {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return fmt.Sprintf("%.1f PB", value)
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d 天 %d 小时", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%d 小时 %d 分钟", hours, minutes)
	}
	return fmt.Sprintf("%d 分钟", minutes)
}
