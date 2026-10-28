package stats_test

import (
	"testing"
	"time"

	"github.com/Milki7/gitpulse/internal/model"
	"github.com/Milki7/gitpulse/internal/stats"
)

func TestSortByStars(t *testing.T) {
	cases := []struct {
		name string
		in   []model.Repo
		want []string // expected order of repo names
	}{
		{
			name: "descending by stars",
			in: []model.Repo{
				{Name: "low", Stars: 5},
				{Name: "high", Stars: 100},
				{Name: "mid", Stars: 50},
			},
			want: []string{"high", "mid", "low"},
		},
		{
			name: "equal stars sorted by name ascending",
			in: []model.Repo{
				{Name: "bravo", Stars: 10},
				{Name: "alpha", Stars: 10},
				{Name: "charlie", Stars: 10},
			},
			want: []string{"alpha", "bravo", "charlie"},
		},
		{
			name: "already sorted",
			in: []model.Repo{
				{Name: "first", Stars: 99},
				{Name: "second", Stars: 50},
			},
			want: []string{"first", "second"},
		},
		{
			name: "single element",
			in: []model.Repo{
				{Name: "only", Stars: 42},
			},
			want: []string{"only"},
		},
		{
			name: "empty slice",
			in:   []model.Repo{},
			want: []string{},
		},
		{
			name: "zero stars mixed with non-zero",
			in: []model.Repo{
				{Name: "zeroed", Stars: 0},
				{Name: "popular", Stars: 200},
				{Name: "niche", Stars: 3},
			},
			want: []string{"popular", "niche", "zeroed"},
		},
		{
			name: "other fields do not affect sort",
			in: []model.Repo{
				{Name: "old", Stars: 10, LastPush: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)},
				{Name: "new", Stars: 10, LastPush: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
			want: []string{"new", "old"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stats.SortByStars(tc.in)

			if len(tc.in) != len(tc.want) {
				t.Fatalf("got %d repos, want %d", len(tc.in), len(tc.want))
			}
			for i, r := range tc.in {
				if r.Name != tc.want[i] {
					t.Errorf("position %d: got %q, want %q", i, r.Name, tc.want[i])
				}
			}
		})
	}
}
