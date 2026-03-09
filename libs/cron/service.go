package cron

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
	"valerygordeev/go/exercises/libs/base"
)

const (
	ServiceVersion = "1.0.0"
)

type CronService struct {
	config       *ServiceConfig
	db           *base.DB
	engine       *inMemoryEngine
	wg           sync.WaitGroup
	mainEvents   chan struct{}
	workerQueues []chan []*CronRecord
}

func NewCronService(db *base.DB, config *ServiceConfig) (*CronService, error) {
	result := &CronService{
		config:       config,
		db:           db,
		engine:       NewInMemoryEngine(config.MaxEventDistance.Duration),
		wg:           sync.WaitGroup{},
		mainEvents:   make(chan struct{}, 100),
		workerQueues: make([]chan []*CronRecord, config.Workers),
	}

	for idx := range config.Workers {
		result.workerQueues[idx] = make(chan []*CronRecord, config.WorkerQueueSize)
		result.wg.Add(1)
		go result.workerRun(idx)
	}

	result.wg.Add(1)
	go result.mainRun()

	return result, nil
}

func (s *CronService) mainRun() {
	log.Printf("mainRun() - start")
	defer s.wg.Done()
	var waitTime = time.Duration(0)
	for true {
		select {
		case <-s.mainEvents:
			log.Printf("mainRun() - finish")
			return
		case <-time.After(waitTime):
		}
		for true {
			before := time.Now().UTC()
			records, err := s.engine.SelectForProcessing(before, s.config.MaxSelectCount)
			recordsCount := len(records)
			log.Printf("selectCronRecords(%v)=%d,%v", before, recordsCount, err)
			if err != nil {
				workersCount := len(s.workerQueues)
				if recordsCount == 0 {
					break
				}
				if recordsCount < workersCount {
					s.workerQueues[0] <- records
				} else {
					recordsCountPerWorker := recordsCount / workersCount
					startIndex := 0
					for idx := range workersCount {
						s.workerQueues[idx] <- records[startIndex : startIndex+recordsCountPerWorker]
						startIndex += recordsCountPerWorker
					}
					if startIndex < recordsCount {
						s.workerQueues[0] <- records[startIndex:]
					}
				}
			} else {
				log.Printf("mainRun() - select=%v", err)
				break
			}
		}
	}
}

func (s *CronService) workerRun(idx int) {
	defer s.wg.Done()
	client := http.Client{}
	log.Printf("workerRun[%d] - start", idx)
	for true {
		records := <-s.workerQueues[idx]
		if records == nil {
			log.Printf("workerRun[%d] - finish", idx)
			return
		}
		for _, rec := range records {
			resp, err := client.Get(rec.WebHook)
			if err == nil {
				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
					err = fmt.Errorf("[%s] return %d", rec.WebHook, resp.StatusCode)
				}
			}
			s.engine.FinishRecordProcessing(rec.ID, err)
		}
	}
}

func (s *CronService) Close() {
	for _, wq := range s.workerQueues {
		wq <- nil
	}
	s.mainEvents <- struct{}{}
	s.wg.Wait()
}

func (s *CronService) List(query Query) ([]CronRecord, error) {
	return nil, base.ErrorNotImplemented
}

func (s *CronService) Create(id string, at time.Time, webHook string) error {
	_, err := url.Parse(webHook)
	if err != nil {
		return err
	}
	return s.engine.InsertCronRecord(id, at, webHook)
}

func (s *CronService) Abort(id string) error {
	s.engine.FinishRecordProcessing(id, base.ErrorAborted)
	return nil
}
