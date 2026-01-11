package models

import "time"

type UserSearch struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Keyword       string    `json:"keyword"`
	Country       string    `json:"country"`
	Location      string    `json:"location"`
	Language      string    `json:"language"`
	Frequency     string    `json:"frequency"`
	HoursOld      int       `json:"hours_old"`
	Exclude       string    `json:"exclude"`
	ResultsWanted int       `json:"results_wanted"`
	LastRun       time.Time `json:"last_run"`
}

type CreateSearchRequest struct {
	Keyword       string `json:"keyword"`
	Country       string `json:"country"`
	Location      string `json:"location"`
	Language      string `json:"language"`
	Frequency     string `json:"frequency"` // optional, default hourly
	HoursOld      int    `json:"hours_old"`
	Exclude       string `json:"exclude"`
	ResultsWanted int    `json:"results_wanted"`
}
