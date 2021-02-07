package helpers

import (
	"os"
	"strconv"
	"strings"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/houstonheat/docker-janitor/pkg/types"
	log "github.com/sirupsen/logrus"
)

type imageParts struct {
	ImageName string
	ImageTag  string
}

// GetEnv OS lookup wrapper for string ENVs
// returns default value if given ENV not exists
func GetEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}

// GetEnvBool OS lookup wrapper for boolean ENVs
// returns default value if given ENV not exists
func GetEnvBool(key string, defaultVal bool) bool {
	if envVal, ok := os.LookupEnv(key); ok {
		envBool, err := strconv.ParseBool(envVal)
		if err == nil {
			return envBool
		}
	}
	return defaultVal
}

// IsFiltered checks if given image is excluded
// fullname: my.repo.com/ki/base/system:latest
// image_name: 	 my.repo.com/ki/base/system
// image_tag: 	 latest
func IsFiltered(filters types.Filters, image dockerTypes.ImageSummary) []types.Tag {
	var tagsToDelete []types.Tag
	repoTags := make(map[string]imageParts)

	for _, tag := range image.RepoTags {
		if i := strings.LastIndex(tag, ":"); i >= 0 {
			log.Debugf("Image %s has parts: %s, %s", tag, tag[:i], tag[i+1:])
			repoTags[tag] = imageParts{ImageName: tag[:i], ImageTag: tag[i+1:]}
		} else {
			log.Debugf("Tag %s, has invalid format", tag)
		}
	}

	for tag, parts := range repoTags {
		skip := false
		if _, ok := filters.FullnameFilter[tag]; ok {
			log.Debugf("Image \"%s\" was fully filtered", tag)
			skip = true
		}
		if skip {
			continue
		}

		for filter := range filters.NameFilter {
			if strings.Contains(parts.ImageName, filter) {
				log.Debugf("Name \"%s\" was filtered with \"%s\" filter", parts.ImageName, filter)
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		for filter := range filters.TagFilter {
			if strings.Contains(parts.ImageTag, filter) {
				log.Debugf("Tag \"%s\" was filtered with \"%s\" filter", parts.ImageTag, filter)
				skip = true
				break
			}
		}

		if skip {
			continue
		} else {
			tagsToDelete = append(tagsToDelete, types.Tag{ImageID: image.ID, ImageTag: tag, ImageSize: image.Size})
		}
	}

	return tagsToDelete
}

// PrintResult prints cleaner results info
// TODO Replace this method with metrics
func PrintResult(totalSpaceReduced int64, intervalSpaceReduced int64, count int) {
	//sizeInMb := float64(totalSpaceReduced) / 1000 / 1000
	log.Infof("Total cleaned from start: %.2fMB; iteration cleaned: %.2fMB; deleted count: %d", float64(totalSpaceReduced)/1000/1000, float64(intervalSpaceReduced)/1000/1000, count)
}
