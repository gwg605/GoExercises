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
	OwnerID string    `json:"own"`
}

type Query struct {
	OwnerID string
	Limit   int
	Before  time.Time
	After   time.Time
}

type CronAPI interface {
	List(query Query) ([]CronRecord, error)
	Create(at time.Time, webHook string, ownerID string) (string, error)
	Abort(id string) error
}
