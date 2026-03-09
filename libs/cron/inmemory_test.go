package cron

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"valerygordeev/go/exercises/libs/base"

	"github.com/google/uuid"
)

func TestCronItemsList(t *testing.T) {
	list0 := cronItemsList{}
	list1 := cronItemsList{}

	i0 := &cronItem{Value: CronRecord{ID: "0"}}
	list0.AddTail(i0, i0)
	if list0.Head != i0 || list0.Tail != i0 {
		t.Fatalf("list0.AddTail(i0) failed")
	}

	i1 := &cronItem{Value: CronRecord{ID: "1"}}
	list0.AddTail(i1, i1)
	if list0.Head != i0 || list0.Tail != i1 {
		t.Fatalf("list0.AddTail(i1) failed")
	}
	if i0.Prev != nil || i0.Next != i1 {
		t.Fatalf("Wrong i0")
	}
	if i1.Prev != i0 || i1.Next != nil {
		t.Fatalf("Wrong i1")
	}

	i2 := &cronItem{Value: CronRecord{ID: "2"}}
	list0.AddTail(i2, i2)
	if list0.Head != i0 || list0.Tail != i2 {
		t.Fatalf("list0.AddTail(i2) failed")
	}
	if i0.Prev != nil || i0.Next != i1 {
		t.Fatalf("Wrong i0")
	}
	if i1.Prev != i0 || i1.Next != i2 {
		t.Fatalf("Wrong i1")
	}
	if i2.Prev != i1 || i2.Next != nil {
		t.Fatalf("Wrong i2")
	}

	list0.SetHead(i2)
	if list0.Head != i2 || list0.Tail != i2 {
		t.Fatalf("SetHead(i2) failed")
	}
	if i0.Prev != nil || i0.Next != i1 {
		t.Fatalf("Wrong i0")
	}
	if i1.Prev != i0 || i1.Next != nil {
		t.Fatalf("Wrong i1")
	}
	if i2.Prev != nil || i2.Next != nil {
		t.Fatalf("Wrong i2")
	}

	list1.AddTail(i0, i1)
	if list1.Head != i0 || list1.Tail != i1 {
		t.Fatalf("list1.AddTail(i0) failed")
	}
	if i0.Prev != nil || i0.Next != i1 {
		t.Fatalf("Wrong i0")
	}
	if i1.Prev != i0 || i1.Next != nil {
		t.Fatalf("Wrong i1")
	}
	if list0.Head != i2 || list0.Tail != i2 {
		t.Fatalf("SetHead(i2) failed")
	}
	if i2.Prev != nil || i2.Next != nil {
		t.Fatalf("Wrong i2")
	}

	list0.SetHead(nil)
	if list0.Head != nil || list0.Tail != nil {
		t.Fatalf("SetHead(nil) failed")
	}
	if i0.Prev != nil || i0.Next != i1 {
		t.Fatalf("Wrong i0")
	}
	if i1.Prev != i0 || i1.Next != nil {
		t.Fatalf("Wrong i1")
	}
	if i2.Prev != nil || i2.Next != nil {
		t.Fatalf("Wrong i2")
	}

	list1.AddTail(i2, i2)
	if list1.Head != i0 || list1.Tail != i2 {
		t.Fatalf("list1.AddTail(i2, i2) failed")
	}
	if i0.Prev != nil || i0.Next != i1 {
		t.Fatalf("Wrong i0")
	}
	if i1.Prev != i0 || i1.Next != i2 {
		t.Fatalf("Wrong i1")
	}
	if i2.Prev != i1 || i2.Next != nil {
		t.Fatalf("Wrong i2")
	}

	i3 := &cronItem{Value: CronRecord{ID: "3"}}
	list0.AddTail(i3, i3)
	if list0.Head != i3 || list0.Tail != i3 {
		t.Fatalf("list0.AddTail(i3, i3) failed")
	}
	if i3.Prev != nil || i3.Next != nil {
		t.Fatalf("Wrong i3")
	}

	b1 := list1.Remove(i1)
	if b1 != true {
		t.Fatalf("list1.Remove(i1) failed")
	}
	if list1.Head != i0 || list1.Tail != i2 {
		t.Fatalf("Wrong list1")
	}
	if i1.Prev != nil || i1.Next != nil {
		t.Fatalf("Wrong i1")
	}
	if i0.Prev != nil || i0.Next != i2 {
		t.Fatalf("Wrong i0")
	}
	if i2.Prev != i0 || i2.Next != nil {
		t.Fatalf("Wrong i2")
	}

	b3 := list1.Remove(i3)
	if b3 != false {
		t.Fatalf("list1.Remove(i1) failed")
	}
	if list1.Head != i0 || list1.Tail != i2 {
		t.Fatalf("Wrong list1")
	}
	if list0.Head != i3 || list0.Tail != i3 {
		t.Fatalf("Wrong list0")
	}

	b3 = list0.Remove(i3)
	if b3 != true {
		t.Fatalf("list0.Remove(i3) failed")
	}
	if list0.Head != nil || list0.Tail != nil {
		t.Fatalf("Wrong list0")
	}
	if list1.Head != i0 || list1.Tail != i2 {
		t.Fatalf("Wrong list1")
	}
	if i3.Prev != nil || i3.Next != nil {
		t.Fatalf("Wrong i3")
	}

	b2 := list1.Remove(i2)
	if b2 != true {
		t.Fatalf("list1.Remove(i2) failed")
	}
	if list1.Head != i0 || list1.Tail != i0 {
		t.Fatalf("Wrong list1")
	}
	if i2.Prev != nil || i2.Next != nil {
		t.Fatalf("Wrong i2")
	}
	if i0.Prev != nil || i0.Next != nil {
		t.Fatalf("Wrong i0")
	}

	b0 := list1.Remove(i0)
	if b0 != true {
		t.Fatalf("list1.Remove(i0) failed")
	}
	if list1.Head != nil || list1.Tail != nil {
		t.Fatalf("Wrong list1")
	}
	if i0.Prev != nil || i0.Next != nil {
		t.Fatalf("Wrong i0")
	}
}

func TestInMemoryEngineBasic(t *testing.T) {
	startTime := time.Now().UTC()
	engine := NewInMemoryEngine(time.Hour)
	// engine sets queueStartTime dynamically, but test requires known start time for building correct timeline and check results
	engine.queueStartTime = startTime.UnixMilli()

	for i := range 100 {
		key := fmt.Sprintf("id#%d", i)
		at := startTime.Add(time.Duration(i+500) * time.Millisecond)
		err := engine.InsertCronRecord(key, at, "http://not.found/"+key)
		if err != nil {
			t.Fatalf("InsertCronRecord()=%v", err)
		}
	}

	records, err := engine.SelectForProcessing(startTime, 10)
	if err != nil {
		t.Fatalf("SelectForProcessing()=%v", err)
	}
	if len(records) != 0 {
		t.Fatalf("Wrong records count %d. Expected 0", len(records))
	}

	for _, rec := range records {
		engine.FinishRecordProcessing(rec.ID, nil)
	}
	if engine.processingQueue.Head != nil || engine.processingQueue.Tail != nil {
		t.Fatalf("Wrong processingQueue")
	}

	records, err = engine.SelectForProcessing(startTime.Add(500*time.Millisecond), 10)
	if err != nil {
		t.Fatalf("SelectForProcessing()=%v", err)
	}
	if len(records) != 10 {
		t.Fatalf("Wrong records count %d. Expected 10", len(records))
	}

	for _, rec := range records {
		engine.FinishRecordProcessing(rec.ID, nil)
	}
	if engine.processingQueue.Head != nil || engine.processingQueue.Tail != nil {
		t.Fatalf("Wrong processingQueue")
	}

	records, err = engine.SelectForProcessing(startTime.Add(510*time.Millisecond), 5)
	if err != nil {
		t.Fatalf("SelectForProcessing()=%v", err)
	}
	if len(records) != 5 {
		t.Fatalf("Wrong records count %d. Expected 5", len(records))
	}

	for _, rec := range records {
		engine.FinishRecordProcessing(rec.ID, nil)
	}
	if engine.processingQueue.Head != nil || engine.processingQueue.Tail != nil {
		t.Fatalf("Wrong processingQueue")
	}

	records, err = engine.SelectForProcessing(startTime.Add(520*time.Millisecond), 20)
	if err != nil {
		t.Fatalf("SelectForProcessing()=%v", err)
	}
	if len(records) != 15 {
		t.Fatalf("Wrong records count %d. Expected 15", len(records))
	}

	for _, rec := range records {
		engine.FinishRecordProcessing(rec.ID, nil)
	}
	if engine.processingQueue.Head != nil || engine.processingQueue.Tail != nil {
		t.Fatalf("Wrong processingQueue")
	}

	for i := range 10 {
		key := fmt.Sprintf("id#%d", i+30)
		engine.FinishRecordProcessing(key, base.ErrorAborted)
	}

	records, err = engine.SelectForProcessing(startTime.Add(530*time.Millisecond), 10)
	if err != nil {
		t.Fatalf("SelectForProcessing()=%v", err)
	}
	if len(records) != 0 {
		t.Fatalf("Wrong records count %d. Expected 0", len(records))
	}

	at := startTime.Add(3601 * time.Second)
	err = engine.InsertCronRecord("out-of-range", at, "http://not.found/")
	if !errors.Is(err, base.ErrorOutOfRange) {
		t.Fatalf("InsertCronRecord()=%v", err)
	}

	err = engine.InsertCronRecord("out-of-range", startTime, "http://not.found/")
	if !errors.Is(err, base.ErrorOutOfRange) {
		t.Fatalf("InsertCronRecord()=%v", err)
	}

	at = startTime.Add(time.Duration(70+500) * time.Millisecond)
	err = engine.InsertCronRecord("id#70", at, "http://not.found/id#70")
	if err != nil {
		t.Fatalf("InsertCronRecord()=%v", err)
	}
}

func TestInMemoryEngineMainLoop(t *testing.T) {
	var stop atomic.Bool
	engine := NewInMemoryEngine(time.Hour)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		limit := 25
		for !stop.Load() {
			for !stop.Load() {
				before := time.Now().UTC()
				records, err := engine.SelectForProcessing(before, limit)
				//log.Printf("SelectForProcessing(%v, %d)=%d, %v", before, limit, len(records), err)
				if err != nil {
					log.Printf("SelectForProcessing(%v, %d)=%v", before, limit, err)
					break
				}
				if len(records) == 0 {
					break
				}
				for _, rec := range records {
					engine.FinishRecordProcessing(rec.ID, nil)
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	for i := range 256 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var iterations int
			startTime := time.Now()
			for !stop.Load() {
				u, err := uuid.NewV7()
				if err != nil {
					log.Printf("[%d] uuid.NewV7()()=%v", i, err)
				} else {
					shift := time.Duration(rand.IntN(1000)) * time.Millisecond
					id := u.String()
					at := time.Now().UTC().Add(shift)
					webHook := "http://host/webhook/" + id
					err = engine.InsertCronRecord(id, at, webHook)
					if err != nil {
						log.Printf("[%d] InsertCronRecord(%s, %v)=%v", i, id, at, err)
					}
				}
				iterations++

				time.Sleep(100 * time.Microsecond)
			}
			duration := time.Since(startTime)
			rps := float64(iterations) / duration.Seconds()
			log.Printf("[%d] Iterations=%d, time=%v, rps=%.03f", i, iterations, duration, rps)
		}()
	}

	time.Sleep(10 * time.Second)

	stop.Store(true)
	wg.Wait()
}

func BenchmarkInMemoryBasic(b *testing.B) {
	startTime := time.Now().UTC()
	engine := NewInMemoryEngine(time.Hour)

	for i := range 300000 {
		key := fmt.Sprintf("id#%d", i)
		at := startTime.Add(time.Duration(i+500) * time.Millisecond)
		err := engine.InsertCronRecord(key, at, "http://not.found/"+key)
		if err != nil {
			b.Fatalf("InsertCronRecord()=%v", err)
		}
	}

	b.ResetTimer()

	var totalRecords int
	curTime := startTime.Add(515 * time.Millisecond)
	for i := range b.N {
		_ = i
		records, err := engine.SelectForProcessing(curTime, 15)
		//log.Printf("records=%d, time=+%d (%v), total=%d", len(records), curTime.UnixMilli()-engine.queueStartTime, curTime, totalRecords)
		if err != nil {
			b.Fatalf("SelectForProcessing()=%v", err)
		}

		for _, rec := range records {
			engine.FinishRecordProcessing(rec.ID, nil)
		}
		if engine.processingQueue.Head != nil || engine.processingQueue.Tail != nil {
			b.Fatalf("Wrong processingQueue")
		}
		totalRecords += len(records)

		shift := len(records)
		if shift == 0 {
			shift = 10
		}
		curTime = curTime.Add(time.Duration(shift) * time.Millisecond)
	}
}
