package model

import "time"

// Repo holds the fields gitpulse displays for a GitHub repository.
type Repo struct {
	Name     string
	Owner    string
	Language string
	Stars    int
	Forks    int
	LastPush time.Time
	URL      string
}
