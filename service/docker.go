package service

import (
	"context"

	ikuaisdk "github.com/zy84338719/ikuai-api"
	"github.com/zy84338719/ikuai-api/types"
)

type dockerService struct {
	client *ikuaisdk.Client
}

func NewDockerService(client *ikuaisdk.Client) DockerService {
	return &dockerService{client: client}
}

func (s *dockerService) GetImages(ctx context.Context) ([]types.DockerImageItem, error) {
	var resp types.DockerImageShowResponse
	if err := s.client.Call(ctx, "docker_image", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *dockerService) GetContainers(ctx context.Context) ([]types.DockerContainerItem, error) {
	var resp types.DockerContainerShowResponse
	if err := s.client.Call(ctx, "docker_container", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *dockerService) GetNetworks(ctx context.Context) ([]types.DockerNetworkItem, error) {
	var resp types.DockerNetworkShowResponse
	if err := s.client.Call(ctx, "docker_network", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}

func (s *dockerService) GetComposes(ctx context.Context) ([]types.DockerComposeItem, error) {
	var resp types.DockerComposeShowResponse
	if err := s.client.Call(ctx, "docker_compose", "show", nil, &resp); err != nil {
		return nil, err
	}
	return resp.GetData(), nil
}
