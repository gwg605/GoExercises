package cron

import (
	"sync"
	"time"
	"valerygordeev/go/exercises/libs/base"
)

const (
	ServiceVersion = "1.0.0"
)

type CronService struct {
	config  *ServiceConfig
	db      *base.DB
	running bool
	wg      sync.WaitGroup
}

func NewCronService(db *base.DB, config *ServiceConfig) (*CronService, error) {
	result := &CronService{
		config:  config,
		db:      db,
		running: true,
		wg:      sync.WaitGroup{},
	}
	return result, nil
}

func (s *CronService) Close() {
	s.running = false
	s.wg.Wait()
}

func (s *CronService) List(query Query) ([]CronRecord, error) {
	return nil, base.ErrorNotImplemented
}

func (s *CronService) Create(at time.Time, webHook string, ownerID string) (string, error) {
	return base.NilString, base.ErrorNotImplemented
}

func (s *CronService) Abort(id string) error {
	return base.ErrorNotImplemented
}
