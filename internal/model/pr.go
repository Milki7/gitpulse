package model

import "time"

// PR represents an open pull request in gitpulse.
type PR struct {
	Number    int
	Title     string
	Repo      string
	Author    string
	Role      string
	CreatedAt time.Time
	URL       string
}
