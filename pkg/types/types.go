package types

import (
	"time"
)

// Filter stores objects (names, tags) of docker images to exclude
type Filter map[string]struct{}

// Filters stores map avaliable filters
type Filters struct {
	FullnameFilter Filter
	NameFilter     Filter
	TagFilter      Filter
}

// Options defines cleaner options
type Options struct {

	// Do not change anything
	DryRun bool

	// Enable unsued containters cleaning
	ClearContainers bool

	// Enable unused images cleaning
	ClearNetworks bool

	// Enable unsued volumes cleaning
	ClearVolumes bool

	// Enable unused images cleaning
	ClearImages bool

	// The time period for which images should be saved
	Freshness time.Duration

	// Map of pattern to save images by
	Filters Filters

	// Triggers filters check or permanent deletion
	FiltersExists bool

	// Triggers freshness check or permanent images prune
	FreshnessExists bool
}

// Tag custom type to separate each tag from it's image
type Tag struct {
	ImageID   string
	ImageTag  string
	ImageSize int64
}
