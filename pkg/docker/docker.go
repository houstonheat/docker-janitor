package docker

import (
	"context"
	"strconv"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/houstonheat/docker-janitor/pkg/types"
	log "github.com/sirupsen/logrus"
)

// ListDockerImages lists all message on the given docker client
func ListDockerImages(cli *client.Client) []dockerTypes.ImageSummary {
	images, err := cli.ImageList(context.Background(), dockerTypes.ImageListOptions{All: true})
	if err != nil {
		log.Errorf("Failed to list docker images on host, err: %s", err)
	}
	return images
}

// ListDockerContainers lists all message on the given docker client
func ListDockerContainers(cli *client.Client) []dockerTypes.Container {
	containers, err := cli.ContainerList(context.Background(), dockerTypes.ContainerListOptions{All: true})
	if err != nil {
		log.Errorf("Failed to list docker images on host, err: %s", err)
	}
	return containers
}

// DeleteDockerImage removes image by ID or tag by Name
func DeleteDockerImage(cli *client.Client, tag types.Tag) (reducedSpace int64) {
	removeOptions := dockerTypes.ImageRemoveOptions{
		Force:         true,
		PruneChildren: false,
	}

	var image string
	if tag.ImageTag == "<none>:<none>" || tag.ImageTag == "" {
		image = tag.ImageID
	} else {
		image = tag.ImageTag
	}

	deleted, err := cli.ImageRemove(context.Background(), image, removeOptions)
	if err != nil {
		log.Errorf("Can't remove image %s, err: %s", tag.ImageID, err)
	} else {
		for _, del := range deleted {
			if del.Deleted != "" {
				log.Debugf("Deleted: %s; Space reclaimed: %d", del.Deleted, tag.ImageSize)
				reducedSpace = tag.ImageSize
			} else {
				log.Debugf("Untagged: %s", del.Untagged)
			}
		}
	}

	return reducedSpace
}

// DeleteDockerTags removes given tags
func DeleteDockerTags(cli *client.Client, tags []types.Tag) int64 {
	var totalCleanedSize int64
	for _, tag := range tags {
		reducedSpace := DeleteDockerImage(cli, tag)
		totalCleanedSize = totalCleanedSize + reducedSpace
	}
	return totalCleanedSize
}

// PruneUnusedVolumes removes all unsused volumes
// same as 'docker volume prune'
func PruneUnusedVolumes(cli *client.Client) {
	_, err := cli.VolumesPrune(context.Background(), filters.Args{})
	if err != nil {
		log.Errorf("Failed to delete unused volumes, err: %s", err)
	}
}

// PruneUnusedContainers removes all unsused containers
// same as 'docker container prune'
func PruneUnusedContainers(cli *client.Client) {
	_, err := cli.ContainersPrune(context.Background(), filters.Args{})
	if err != nil {
		log.Errorf("Failed to prune unused containers, err: %s", err)
	}
}

// PruneUnusedNetworks removes all unsused networks
// same as 'docker network prune'
func PruneUnusedNetworks(cli *client.Client) {
	_, err := cli.NetworksPrune(context.Background(), filters.Args{})
	if err != nil {
		log.Errorf("Failed to prune unused networks, err: %s", err)
	}
}

// PruneUnusedImages removes all unsused images
// same as 'docker image prune'
func PruneUnusedImages(cli *client.Client, dangling bool) (report dockerTypes.ImagesPruneReport) {
	args := filters.NewArgs()
	args.Add("dangling", strconv.FormatBool(dangling))

	report, err := cli.ImagesPrune(context.Background(), args)
	if err != nil {
		log.Errorf("Failed to prune unused iamges, err: %s", err)
	}
	return
}
