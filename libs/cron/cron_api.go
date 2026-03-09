package cron

import (
	"time"
)

const (
	ServiceShortName = "CRON"
)

type CronRecord struct {
	ID      string    `json:"id"`
	At      time.Time `json:"at"`
	WebHook string    `json:"wh"`
}

type Query struct {
	Limit  int
	Before time.Time
	After  time.Time
}

type CronAPI interface {
	List(query Query) ([]CronRecord, error)
	Create(id string, at time.Time, webHook string) error
	Abort(id string) error
}
