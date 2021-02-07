package helpers

import (
	"testing"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/houstonheat/docker-janitor/pkg/types"
)

func Test_isFiltered(t *testing.T) {
	// Fixtures
	testImageID := "sha256:22667f53682a2920948d19c7133ab1c9c3f745805c14125859d20cede07f11f9"
	testSize := int64(1231733)

	var filtersTests = []struct {
		name    string
		image   dockerTypes.ImageSummary
		filters types.Filters
		want    []types.Tag
	}{
		{
			"Image fullname filters test",
			dockerTypes.ImageSummary{
				ID: testImageID,
				RepoTags: []string{
					"my.repo.org:9000/my/image/name:latest",
					"other.repo.org/important/image:5",
					"other.repo.org/important/image:6",
				},
				Size: testSize,
			},
			types.Filters{
				FullnameFilter: types.Filter{
					"other.repo.org/important/image:5": struct{}{}},
			},
			[]types.Tag{
				{
					ImageID:   testImageID,
					ImageSize: testSize,
					ImageTag:  "my.repo.org:9000/my/image/name:latest",
				},
				{
					ImageID:   testImageID,
					ImageSize: testSize,
					ImageTag:  "other.repo.org/important/image:6",
				},
			},
		},
		{
			"Image name filters test",
			dockerTypes.ImageSummary{
				ID: testImageID,
				RepoTags: []string{
					"myimage:latest",
					"myimage-new:latest",
					"my-awesome-image:latest",
				},
				Size: testSize,
			},
			types.Filters{
				NameFilter: types.Filter{
					"myimage": struct{}{}},
			},
			[]types.Tag{
				{
					ImageID:   testImageID,
					ImageSize: testSize,
					ImageTag:  "my-awesome-image:latest",
				},
			},
		},
		{
			"Image tags filters test",
			dockerTypes.ImageSummary{
				ID: testImageID,
				RepoTags: []string{
					"myimage:stable",
					"myimage-new:stable-twice",
					"my-awesome-image:5.22",
				},
				Size: testSize,
			},
			types.Filters{
				TagFilter: types.Filter{
					"stable": struct{}{}},
			},
			[]types.Tag{
				{
					ImageID:   testImageID,
					ImageSize: testSize,
					ImageTag:  "my-awesome-image:5.22",
				},
			},
		},
	}

	for _, tt := range filtersTests {
		t.Run(tt.name, func(t *testing.T) {
			del := IsFiltered(tt.filters, tt.image)
			if !sameSlice(tt.want, del) {
				t.Errorf("Tags mismatch in\nWANT: %#v\nGOT: %#v", tt.want, del)
			}
		})
	}

}

func sameSlice(x, y []types.Tag) bool {
	if len(x) != len(y) {
		return false
	}

	diff := make(map[types.Tag]int, len(x))
	for _, _x := range x {
		diff[_x]++
	}
	for _, _y := range y {
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y]--
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	if len(diff) == 0 {
		return true
	}
	return false
}
