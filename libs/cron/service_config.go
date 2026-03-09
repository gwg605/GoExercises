package cron

import (
	"time"
	"valerygordeev/go/exercises/libs/base"
)

type ServiceConfig struct {
	Workers          int           `json:"workers"`
	WorkerQueueSize  int           `json:"worker_queue_size"`
	MaxEventDistance base.Duration `json:"max_event_distance"`
	MaxSelectCount   int           `json:"max_select_count"`
}

func NewServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		Workers:          4,
		WorkerQueueSize:  100,
		MaxEventDistance: base.Duration{Duration: 12 * time.Hour},
		MaxSelectCount:   10,
	}
}
