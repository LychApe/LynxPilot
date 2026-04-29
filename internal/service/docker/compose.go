package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

type ComposeProject struct {
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Running   int      `json:"running"`
	Stopped   int      `json:"stopped"`
	Services  []string `json:"services"`
	Networks  []string `json:"networks"`
	ConfigDir string   `json:"config_dir,omitempty"`
}

func dockerComposeEnv() []string {
	cfg := loadConnectionConfig()
	env := os.Environ()

	if cfg == nil || cfg.Host == "" {
		return env
	}

	env = append(env, "DOCKER_HOST="+cfg.Host)

	if cfg.TLSVerify {
		env = append(env, "DOCKER_TLS_VERIFY=1")
	} else {
		for i, e := range env {
			if strings.HasPrefix(e, "DOCKER_TLS_VERIFY=") {
				env[i] = "DOCKER_TLS_VERIFY="
			}
		}
	}

	if cfg.CertPath != "" {
		env = append(env, "DOCKER_CERT_PATH="+cfg.CertPath)
	}

	return env
}

func newDockerCmd(args ...string) *exec.Cmd {
	cmd := exec.Command("docker", args...)
	cmd.Env = dockerComposeEnv()
	return cmd
}

func getComposeContainers(projectName string) ([]container.Summary, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	args := filters.NewArgs()
	args.Add("label", "com.docker.compose.project="+projectName)

	return cli.ContainerList(context.Background(), container.ListOptions{All: true, Filters: args})
}

func ListComposeProjects() ([]ComposeProject, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	args := filters.NewArgs()
	args.Add("label", "com.docker.compose.project")

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true, Filters: args})
	if err != nil {
		return nil, fmt.Errorf("获取 Compose 项目列表失败: %w", err)
	}

	projectMap := make(map[string]*ComposeProject)
	for _, c := range containers {
		projectName := c.Labels["com.docker.compose.project"]
		if projectName == "" {
			continue
		}

		project, ok := projectMap[projectName]
		if !ok {
			project = &ComposeProject{Name: projectName}
			projectMap[projectName] = project
		}

		serviceName := c.Labels["com.docker.compose.service"]
		if serviceName != "" {
			found := false
			for _, s := range project.Services {
				if s == serviceName {
					found = true
					break
				}
			}
			if !found {
				project.Services = append(project.Services, serviceName)
			}
		}

		networkNames := c.Labels["com.docker.compose.network"]
		if networkNames != "" {
			for _, n := range strings.Split(networkNames, ",") {
				n = strings.TrimSpace(n)
				if n == "" {
					continue
				}
				found := false
				for _, existing := range project.Networks {
					if existing == n {
						found = true
						break
					}
				}
				if !found {
					project.Networks = append(project.Networks, n)
				}
			}
		}

		if c.State == "running" {
			project.Running++
		} else {
			project.Stopped++
		}
	}

	if len(projectMap) == 0 {
		return []ComposeProject{}, nil
	}

	result := make([]ComposeProject, 0, len(projectMap))
	for _, p := range projectMap {
		total := p.Running + p.Stopped
		if p.Running == total {
			p.Status = "running"
		} else if p.Running == 0 {
			p.Status = "stopped"
		} else {
			p.Status = "partial"
		}
		sort.Strings(p.Services)
		sort.Strings(p.Networks)
		result = append(result, *p)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func ComposeUpFromContent(content string, projectName string) error {
	cmd := newDockerCmd("compose")
	if projectName != "" {
		cmd = newDockerCmd("compose", "-p", projectName)
	}
	cmd.Args = append(cmd.Args, "-f", "-", "up", "-d")
	cmd.Stdin = strings.NewReader(content)
	output, err := cmd.CombinedOutput()
	if err != nil {
		hint := ""
		if strings.Contains(err.Error(), "executable file not found") || strings.Contains(err.Error(), "not recognized") {
			hint = "\n提示: 本机未安装 docker CLI，compose up 需要在 LynxPilot 所在机器上安装 Docker CLI"
		}
		return fmt.Errorf("Compose up 失败: %s%s: %w", string(output), hint, err)
	}
	return nil
}

func ComposeDown(projectName string, removeVolumes bool) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	containers, err := getComposeContainers(projectName)
	if err != nil {
		return fmt.Errorf("获取项目容器失败: %w", err)
	}

	for _, c := range containers {
		if c.State == "running" {
			timeout := 10
			if err := cli.ContainerStop(context.Background(), c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
				return fmt.Errorf("停止容器 %s 失败: %w", c.ID[:12], err)
			}
		}
		if err := cli.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{Force: true}); err != nil {
			return fmt.Errorf("删除容器 %s 失败: %w", c.ID[:12], err)
		}
	}

	networkArgs := filters.NewArgs()
	networkArgs.Add("label", "com.docker.compose.project="+projectName)
	networks, err := cli.NetworkList(context.Background(), network.ListOptions{Filters: networkArgs})
	if err == nil {
		for _, n := range networks {
			_ = cli.NetworkRemove(context.Background(), n.ID)
		}
	}

	return nil
}

func ComposeStop(projectName string) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	containers, err := getComposeContainers(projectName)
	if err != nil {
		return err
	}

	for _, c := range containers {
		if c.State == "running" {
			timeout := 10
			if err := cli.ContainerStop(context.Background(), c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
				return fmt.Errorf("停止容器 %s 失败: %w", c.ID[:12], err)
			}
		}
	}
	return nil
}

func ComposeStart(projectName string) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	containers, err := getComposeContainers(projectName)
	if err != nil {
		return err
	}

	for _, c := range containers {
		if c.State != "running" {
			if err := cli.ContainerStart(context.Background(), c.ID, container.StartOptions{}); err != nil {
				return fmt.Errorf("启动容器 %s 失败: %w", c.ID[:12], err)
			}
		}
	}
	return nil
}

func ComposeRestart(projectName string) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	containers, err := getComposeContainers(projectName)
	if err != nil {
		return err
	}

	for _, c := range containers {
		timeout := 10
		if err := cli.ContainerRestart(context.Background(), c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
			return fmt.Errorf("重启容器 %s 失败: %w", c.ID[:12], err)
		}
	}
	return nil
}

func ComposeLogs(projectName string, tail string) (string, error) {
	cli, err := getClient()
	if err != nil {
		return "", err
	}
	defer cli.Close()

	if tail == "" {
		tail = "100"
	}

	containers, err := getComposeContainers(projectName)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, c := range containers {
		serviceName := c.Labels["com.docker.compose.service"]
		out, err := cli.ContainerLogs(context.Background(), c.ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Tail:       tail,
			Timestamps: true,
		})
		if err != nil {
			sb.WriteString(fmt.Sprintf("--- [%s] 获取日志失败: %v ---\n", serviceName, err))
			continue
		}

		body, err := readLogOutput(out)
		out.Close()
		if err != nil {
			sb.WriteString(fmt.Sprintf("--- [%s] 读取日志失败: %v ---\n", serviceName, err))
			continue
		}

		if len(body) > 0 {
			sb.WriteString(fmt.Sprintf("--- [%s] ---\n", serviceName))
			sb.WriteString(body)
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

func ComposePs(projectName string) ([]ContainerInfo, error) {
	containers, err := getComposeContainers(projectName)
	if err != nil {
		return nil, err
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

func readLogOutput(rc io.ReadCloser) (string, error) {
	body, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func CheckComposeAvailable() bool {
	cli, err := getClient()
	if err != nil {
		return false
	}
	defer cli.Close()

	_, err = cli.Ping(context.Background())
	return err == nil
}
