package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "net/http/pprof"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/houstonheat/docker-janitor/pkg/docker"
	"github.com/houstonheat/docker-janitor/pkg/helpers"
	"github.com/houstonheat/docker-janitor/pkg/types"
	log "github.com/sirupsen/logrus"
)

var (
	version         = "dev"
	commit          = "unknown"
	date            = "unknown"
	builtBy         = "unknown"
	filtersExists   bool
	freshnessExists bool
)

type customFormatter struct {
	log.TextFormatter
}

func (f *customFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("[%s] [%s] %s\n", entry.Time.Format(f.TimestampFormat), entry.Level.String(), entry.Message)), nil
}

func main() {

	var (
		once             = flag.Bool("once", helpers.GetEnvBool("ONCE", false), "Execute clean task once and exit")
		debug            = flag.Bool("debug", helpers.GetEnvBool("DEBUG", false), "sets log level to debug")
		dryRun           = flag.Bool("dry-run", helpers.GetEnvBool("DRY_RUN", false), "Do not change anything, just print what would be done")
		clearContainers  = flag.Bool("clear-containers", helpers.GetEnvBool("CLEAR_CONTAINERS", false), "Clear unused containers, same as \"docker container prune\"")
		clearNetworks    = flag.Bool("clear-networks", helpers.GetEnvBool("CLEAR_NETWORKS", false), "Clear unused networks, same as \"docker network prune\"")
		clearVolumes     = flag.Bool("clear-volumes", helpers.GetEnvBool("CLEAR_VOLUMES", false), "Clear unused volumes, same as \"docker volume prune\"")
		clearImages      = flag.Bool("clear-images", helpers.GetEnvBool("CLEAR_IMAGES", false), "Clear unused images, same as \"docker image prune -a\"")
		excludeFullnames = flag.String("exclude-fullnames", helpers.GetEnv("EXCLUDE_FULLNAMES", ""), "Comma separated list of images fullnames `(repo[:port]/path:tag[,...])` to exclude from cleaning (e.g. `-exclude-names registry.domain:9000/path/name:v1.0.0`).\nThis option only makes sense when `-clear-images` flag is set")
		excludeNames     = flag.String("exclude-names", helpers.GetEnv("EXCLUDE_NAMES", ""), "Comma separated list of images names `(repo[:port]/path[,...])` to exclude from cleaning (e.g. `-exclude-names registry.domain/path/name,path/name/,ubuntu,myimage`).\nThis option only makes sense when `-clear-images` flag is set")
		excludeTags      = flag.String("exclude-tags", helpers.GetEnv("EXCLUDE_TAGS", ""), "Comma separated list of images tags `(tag[,tag...])` to exclude from cleaning (e.g. `-exclude-tags latest,stable,5.22`).\nThis option only makes sense when `-clear-images` flag is set")
		freshness        = flag.String("freshness", helpers.GetEnv("FRESHNESS", ""), "Freshness will keep images that were created in the given time period. Empty by default")
		interval         = flag.String("interval", helpers.GetEnv("INTERVAL", "12h"), "Interval to check on unused elements")
	)
	flag.Parse()

	// Docker cli SDK requirement
	_, ok := os.LookupEnv("DOCKER_API_VERSION")
	if !ok {
		os.Setenv("DOCKER_API_VERSION", "1.39")
	}

	// Setup loggin options
	log.SetFormatter(&customFormatter{log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.999", FullTimestamp: true, DisableColors: true}})
	if *debug {
		log.SetLevel(log.DebugLevel)
		log.Debugln("Enabling debug output")
	} else {
		log.SetLevel(log.InfoLevel)
	}

	// Fill filters with given parametes
	fullFilters := make(types.Filter)
	nameFilters := make(types.Filter)
	tagsFilters := make(types.Filter)
	log.Debugf("excludeFullnames %#v", *excludeFullnames)
	log.Debugf("excludeNames %#v", *excludeNames)
	log.Debugf("excludeTags %#v", *excludeTags)
	if *excludeNames == "" && *excludeTags == "" && *excludeFullnames == "" {
		log.Debugf("No filters passed, remove all unused images")
		filtersExists = false
	} else {
		log.Debugf("Filters passed")
		filtersExists = true
	}

	if *excludeNames != "" {
		names := strings.Split(*excludeNames, ",")
		for _, f := range names {
			nameFilters[f] = struct{}{}
		}
		log.Debugf("Names filters %#v", nameFilters)
	}

	if *excludeTags != "" {
		tags := strings.Split(*excludeTags, ",")
		for _, f := range tags {
			tagsFilters[f] = struct{}{}
		}
		log.Debugf("Tags filters %#v", tagsFilters)
	}

	if *excludeFullnames != "" {
		fullnames := strings.Split(*excludeFullnames, ",")
		for _, f := range fullnames {
			fullFilters[f] = struct{}{}
		}
		log.Debugf("Fullname filters %#v", fullFilters)
	}

	// Prepare time periods and start app with infinite ticker
	checkInterval, err := time.ParseDuration(*interval)
	if err != nil {
		log.Fatalf("Coudln't parse interval, err: %s", err)
	}

	frashPeriod, err := time.ParseDuration(*freshness)
	if err != nil && *freshness == "" {
		log.Debugf("Freshness isn't set, delete all unused images")
		freshnessExists = false
	} else if err != nil {
		log.Fatalf("Coudln't parse freshness, err: %s", err)
	}
	freshnessExists = true

	if !*clearContainers && !*clearVolumes && !*clearImages && !*clearNetworks {
		log.Info("No cleaner options provided, exit.")
		return
	}

	if *dryRun {
		log.Info("Dry-run option provided, nothing will be deleted")
	}

	options := types.Options{
		DryRun:          *dryRun,
		ClearContainers: *clearContainers,
		ClearVolumes:    *clearVolumes,
		ClearImages:     *clearImages,
		ClearNetworks:   *clearNetworks,
		FiltersExists:   filtersExists,
		Filters: types.Filters{
			NameFilter:     nameFilters,
			FullnameFilter: fullFilters,
			TagFilter:      tagsFilters,
		},
		FreshnessExists: freshnessExists,
		Freshness:       frashPeriod,
	}

	log.Infof("Cleaner started: version %s, commit %s, built at %s by %s", version, commit, date, builtBy)
	log.Infof("Checking unused docker objects every %s, with freshness %s and filters: \"%s\", \"%s\", \"%s\"", checkInterval, frashPeriod, *excludeNames, *excludeTags, *excludeFullnames)
	// We want first tick to happen instantly
	if *once {
		check(options)
		log.Info("Execution complete, exit")
	} else {
		for ; true; <-time.NewTicker(checkInterval).C {
			check(options)
		}
	}

}

func check(opt types.Options) {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Failed to initiate client, err: %s", err)
	}

	if opt.ClearContainers {
		log.Debug("Pruning unused containers")
		if !opt.DryRun {
			docker.PruneUnusedContainers(cli)
		}
	}

	if opt.ClearVolumes {
		log.Debug("Prunning unused volumes")
		if !opt.DryRun {
			docker.PruneUnusedVolumes(cli)
		}
	}

	if opt.ClearNetworks {
		log.Debug("Prunning unused networks")
		if !opt.DryRun {
			docker.PruneUnusedNetworks(cli)
		}
	}

	if opt.ClearImages {
		var totalReducedSpace int64
		var intervalSpaceReduced int64
		var deletedCount int

		// Always remove dangling images, noone really needs them
		if !opt.DryRun {
			report := docker.PruneUnusedImages(cli, true)
			totalReducedSpace = totalReducedSpace + int64(report.SpaceReclaimed)
			intervalSpaceReduced = int64(report.SpaceReclaimed)
			deletedCount = len(report.ImagesDeleted)
		}

		if !opt.FreshnessExists && !opt.FiltersExists {
			log.Debug("Prunning all unused images")
			if !opt.DryRun {
				report := docker.PruneUnusedImages(cli, false)
				totalReducedSpace = totalReducedSpace + int64(report.SpaceReclaimed)
				intervalSpaceReduced = intervalSpaceReduced + int64(report.SpaceReclaimed)
				deletedCount = deletedCount + len(report.ImagesDeleted)
			}
		} else {
			freshImageTime := time.Now().Add(-opt.Freshness)

			// List all images on host
			images := docker.ListDockerImages(cli)
			imageTree := make(map[string]dockerTypes.ImageSummary, len(images))
			for _, image := range images {
				imageTree[image.ID] = image
			}

			// Find all images that used by running containers
			containers := docker.ListDockerContainers(cli)
			used := map[string]string{}
			for _, container := range containers {
				inspected, err := cli.ContainerInspect(context.Background(), container.ID)
				if err != nil {
					log.Debugf("error getting container info for %s: %s", container.ID, err)
					continue
				}

				used[inspected.Image] = container.ID

				parent := imageTree[inspected.Image].ParentID
				for {
					if parent == "" {
						break
					}

					used[parent] = container.ID
					parent = imageTree[parent].ParentID
				}
			}
			log.Debugf("Images that were created in the last %s will be skipped. Images on host: %d", opt.Freshness, len(images))

			for _, image := range images {
				if containerID, ok := used[image.ID]; !ok {
					log.Debugf("Image %v, with tags: %s", image.ID, strings.Join(image.RepoTags, ","))

					// Check if image is older than given FRESHNESS and can be deleted
					if image.Created < freshImageTime.Unix() {

						var tagsToDelete []types.Tag
						// If filters disabled than we can delete whole image by ID with it's tags
						if !opt.FiltersExists {
							tagsToDelete = []types.Tag{
								{
									ImageID:   image.ID,
									ImageSize: image.Size,
								},
							}
						} else {
							tagsToDelete = helpers.IsFiltered(opt.Filters, image)
						}

						if len(tagsToDelete) == 0 {
							log.Debugf("Image %s filttered out and will not be deleted", image.ID)
						} else {
							log.Debugf("Deleting %d tags for image %s", len(tagsToDelete), image.ID)
							if !opt.DryRun {
								reducedSpace := docker.DeleteDockerTags(cli, tagsToDelete)
								totalReducedSpace = totalReducedSpace + reducedSpace
								intervalSpaceReduced = intervalSpaceReduced + reducedSpace
							}
							deletedCount++
						}
					} else {
						log.Debugf("Deletion skipped on image %v: too fresh", image.ID)
					}
				} else {
					log.Debugf("Skip, image %s is used by container %s", image.ID, containerID)
				}
			}

			if intervalSpaceReduced > 0 || deletedCount > 0 {
				helpers.PrintResult(totalReducedSpace, intervalSpaceReduced, deletedCount)
			} else {
				log.Debug("Nothing has been deleted in this iteration")
			}
		}
	}
}
