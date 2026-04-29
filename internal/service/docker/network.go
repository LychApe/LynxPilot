package docker

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/docker/docker/api/types/network"
)

type NetworkInfo struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	IPAMDriver string            `json:"ipam_driver"`
	Subnets    []string          `json:"subnets"`
	Labels     map[string]string `json:"labels"`
	Created    time.Time         `json:"created"`
	Containers []NetworkContainer `json:"containers"`
}

type NetworkContainer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	IPv4  string `json:"ipv4"`
	IPv6  string `json:"ipv6"`
}

type CreateNetworkRequest struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Subnet     string            `json:"subnet"`
	Gateway    string            `json:"gateway"`
	Internal   bool              `json:"internal"`
	Attachable bool              `json:"attachable"`
	Labels     map[string]string `json:"labels"`
}

func ListNetworks() ([]NetworkInfo, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取网络列表失败: %w", err)
	}

	result := make([]NetworkInfo, 0, len(networks))
	for _, n := range networks {
		info := NetworkInfo{
			ID:         n.ID[:12],
			Name:       n.Name,
			Driver:     n.Driver,
			Scope:      n.Scope,
			Internal:   n.Internal,
			Attachable: n.Attachable,
			Labels:     n.Labels,
			Created:    n.Created,
		}

		if n.IPAM.Driver != "" {
			info.IPAMDriver = n.IPAM.Driver
		}

		for _, cfg := range n.IPAM.Config {
			if cfg.Subnet != "" {
				info.Subnets = append(info.Subnets, cfg.Subnet)
			}
		}

		for id, ep := range n.Containers {
			info.Containers = append(info.Containers, NetworkContainer{
				ID:   id[:12],
				Name: ep.Name,
				IPv4: ep.IPv4Address,
				IPv6: ep.IPv6Address,
			})
		}

		result = append(result, info)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func CreateNetwork(req *CreateNetworkRequest) (*NetworkInfo, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	opts := network.CreateOptions{
		Driver:     req.Driver,
		Internal:   req.Internal,
		Attachable: req.Attachable,
		Labels:     req.Labels,
	}

	if opts.Driver == "" {
		opts.Driver = "bridge"
	}

	if req.Subnet != "" {
		ipamConfig := network.IPAMConfig{Subnet: req.Subnet}
		if req.Gateway != "" {
			ipamConfig.Gateway = req.Gateway
		}
		opts.IPAM = &network.IPAM{
			Driver: "default",
			Config: []network.IPAMConfig{ipamConfig},
		}
	}

	resp, err := cli.NetworkCreate(context.Background(), req.Name, opts)
	if err != nil {
		return nil, fmt.Errorf("创建网络失败: %w", err)
	}

	return &NetworkInfo{
		ID:     resp.ID[:12],
		Name:   req.Name,
		Driver: opts.Driver,
	}, nil
}

func RemoveNetwork(networkID string) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	return cli.NetworkRemove(context.Background(), networkID)
}

func InspectNetwork(networkID string) (*NetworkInfo, error) {
	cli, err := getClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	inspect, err := cli.NetworkInspect(context.Background(), networkID, network.InspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("查看网络详情失败: %w", err)
	}

	info := &NetworkInfo{
		ID:         inspect.ID[:12],
		Name:       inspect.Name,
		Driver:     inspect.Driver,
		Scope:      inspect.Scope,
		Internal:   inspect.Internal,
		Attachable: inspect.Attachable,
		Labels:     inspect.Labels,
		Created:    inspect.Created,
	}

	if inspect.IPAM.Driver != "" {
		info.IPAMDriver = inspect.IPAM.Driver
	}

	for _, cfg := range inspect.IPAM.Config {
		if cfg.Subnet != "" {
			info.Subnets = append(info.Subnets, cfg.Subnet)
		}
	}

	for id, ep := range inspect.Containers {
		info.Containers = append(info.Containers, NetworkContainer{
			ID:   id[:12],
			Name: ep.Name,
			IPv4: ep.IPv4Address,
			IPv6: ep.IPv6Address,
		})
	}

	return info, nil
}

func ConnectContainer(networkID string, containerID string) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	return cli.NetworkConnect(context.Background(), networkID, containerID, &network.EndpointSettings{})
}

func DisconnectContainer(networkID string, containerID string, force bool) error {
	cli, err := getClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	return cli.NetworkDisconnect(context.Background(), networkID, containerID, force)
}
