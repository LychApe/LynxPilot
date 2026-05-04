package docker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

const defaultDaemonJSONPath = "/etc/docker/daemon.json"

type MirrorConfig struct {
	URL string `json:"url"`
}

var (
	daemonJSONPath = defaultDaemonJSONPath
	mirrorMu       sync.Mutex
)

func SetDaemonJSONPath(p string) {
	mirrorMu.Lock()
	defer mirrorMu.Unlock()
	daemonJSONPath = p
}

func getDaemonJSONPath() string {
	mirrorMu.Lock()
	defer mirrorMu.Unlock()
	return daemonJSONPath
}

type daemonJSON struct {
	RegistryMirrors []string `json:"registry-mirrors,omitempty"`
}

func GetRegistryMirrors() ([]MirrorConfig, error) {
	path := getDaemonJSONPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []MirrorConfig{}, nil
		}
		return nil, fmt.Errorf("读取 daemon.json 失败: %w", err)
	}

	var cfg daemonJSON
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析 daemon.json 失败: %w", err)
	}

	mirrors := make([]MirrorConfig, 0, len(cfg.RegistryMirrors))
	for _, u := range cfg.RegistryMirrors {
		mirrors = append(mirrors, MirrorConfig{URL: u})
	}
	return mirrors, nil
}

func SaveRegistryMirrors(mirrors []MirrorConfig) error {
	path := getDaemonJSONPath()

	unique := make(map[string]struct{}, len(mirrors))
	urls := make([]string, 0, len(mirrors))
	for _, m := range mirrors {
		if m.URL == "" {
			continue
		}
		if _, exists := unique[m.URL]; exists {
			continue
		}
		unique[m.URL] = struct{}{}
		urls = append(urls, m.URL)
	}
	sort.Strings(urls)

	var cfg daemonJSON
	data, err := os.ReadFile(path)
	if err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &cfg)
	}

	cfg.RegistryMirrors = urls

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 daemon.json 失败: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("替换 daemon.json 失败: %w", err)
	}

	return nil
}
