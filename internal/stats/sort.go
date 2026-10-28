package stats

import (
	"cmp"
	"slices"

	"github.com/Milki7/gitpulse/internal/model"
)

// SortByStars sorts repos by star count descending. Repos with equal
// stars are ordered by name ascending for stable output.
func SortByStars(repos []model.Repo) {
	slices.SortStableFunc(repos, func(a, b model.Repo) int {
		if c := cmp.Compare(b.Stars, a.Stars); c != 0 {
			return c
		}
		return cmp.Compare(a.Name, b.Name)
	})
}

