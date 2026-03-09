package cron

import (
	"sync"
	"time"
	"valerygordeev/go/exercises/libs/base"
)

const (
	timeSlotGranularity = 10
)

type cronItem struct {
	Value CronRecord
	Next  *cronItem
	Prev  *cronItem
}

type cronItemsList struct {
	Head *cronItem
	Tail *cronItem
}

func (l *cronItemsList) AddTail(start *cronItem, last *cronItem) {
	if l.Tail != nil {
		l.Tail.Next = start
	} else {
		l.Head = start
	}
	start.Prev = l.Tail
	l.Tail = last
}

func (l *cronItemsList) SetHead(item *cronItem) {
	l.Head = item
	if l.Head == nil {
		l.Tail = nil
	} else if l.Head.Prev != nil {
		if l.Head.Prev.Next != nil {
			l.Head.Prev.Next = nil
		}
		l.Head.Prev = nil
	}
}

func (l *cronItemsList) Remove(item *cronItem) bool {
	if item.Prev != nil && item.Next != nil {
		item.Prev.Next = item.Next
		item.Next.Prev = item.Prev
	} else {
		if item.Prev == nil && l.Head != item {
			// different list
			return false
		}
		if item.Next == nil && l.Tail != item {
			// different list
			return false
		}
		if item.Prev == nil {
			l.Head = item.Next
		} else {
			item.Prev.Next = item.Next
		}
		if item.Next == nil {
			l.Tail = item.Prev
		} else {
			item.Next.Prev = item.Prev
		}
	}
	item.Next = nil
	item.Prev = nil
	return true
}

type inMemoryEngine struct {
	queueLock       sync.Mutex
	queue           []cronItemsList
	queueStartTime  int64
	queueStartIndex int
	processingQueue cronItemsList
	indexID         map[string]*cronItem
}

func NewInMemoryEngine(maxEventDistance time.Duration) *inMemoryEngine {
	queueSize := int(maxEventDistance.Milliseconds() / timeSlotGranularity)
	return &inMemoryEngine{
		queueLock:       sync.Mutex{},
		queue:           make([]cronItemsList, queueSize),
		queueStartTime:  0,
		queueStartIndex: 0,
		indexID:         make(map[string]*cronItem),
	}
}

func (s *inMemoryEngine) InsertCronRecord(id string, at time.Time, webHook string) error {
	item := &cronItem{Value: CronRecord{ID: id, At: at, WebHook: webHook}}
	s.queueLock.Lock()
	defer s.queueLock.Unlock()

	if s.queueStartTime == 0 {
		s.queueStartTime = time.Now().UnixMilli()
	}

	oldItem, found := s.indexID[id]
	if found {
		if oldItem.Value.At.Equal(at) && oldItem.Value.WebHook == webHook {
			// duplicate
			return nil
		}
		return base.ErrorAlreadyExists
	}

	shift := at.UnixMilli() - s.queueStartTime
	if shift < 0 {
		return base.ErrorOutOfRange
	}

	insertIndex := int(shift) / timeSlotGranularity
	if (insertIndex + 2) >= len(s.queue) {
		return base.ErrorOutOfRange
	}
	insertIndex += s.queueStartIndex

	s.queue[insertIndex%len(s.queue)].AddTail(item, item)
	s.indexID[id] = item
	//log.Printf("[%s] shift=%d, index=%d, count=%d", id, shift, insertIndex, s.queue[insertIndex%len(s.queue)].Count)

	return nil
}

func (s *inMemoryEngine) SelectForProcessing(before time.Time, limit int) ([]*CronRecord, error) {
	result := make([]*CronRecord, 0, limit)
	s.queueLock.Lock()
	defer s.queueLock.Unlock()

	if s.queueStartTime == 0 {
		return result, nil
	}

	shift := before.UnixMilli() - s.queueStartTime
	stopIndex := int(shift)/timeSlotGranularity + s.queueStartIndex
	curIndex := s.queueStartIndex
	for curIndex <= stopIndex {
		curList := &s.queue[curIndex%len(s.queue)]
		startItem := curList.Head
		if startItem != nil {
			var prevItem *cronItem
			curItem := startItem
			for curItem != nil && limit > 0 {
				result = append(result, &curItem.Value)
				limit--
				prevItem = curItem
				curItem = curItem.Next
			}
			if prevItem != nil && prevItem.Next != nil {
				prevItem.Next = nil
			}
			curList.SetHead(curItem)
			s.processingQueue.AddTail(startItem, prevItem)
			if limit == 0 {
				break
			}
		}
		curIndex++

		if (curIndex - s.queueStartIndex) > 2 {
			// move start head on empty list
			s.queueStartIndex++
			s.queueStartTime += timeSlotGranularity
		}
	}

	if s.queueStartIndex >= len(s.queue) {
		s.queueStartIndex %= len(s.queue)
	}

	return result, nil
}

func (s *inMemoryEngine) FinishRecordProcessing(id string, err error) {
	s.queueLock.Lock()
	defer s.queueLock.Unlock()

	item, found := s.indexID[id]
	if !found {
		return
	}
	removed := s.processingQueue.Remove(item)
	if !removed {
		shift := item.Value.At.UnixMilli() - s.queueStartTime
		index := int(shift)/timeSlotGranularity + s.queueStartIndex
		_ = s.queue[index].Remove(item)
	}
	delete(s.indexID, id)
}
