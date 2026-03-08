package trace_test

import (
	"log"
	"math/rand/v2"
	"os"
	"runtime"
	"runtime/trace"
	"sync"
	"sync/atomic"
	"testing"
)

func TestTrace0(t *testing.T) {
	f, err := os.Create("trace0.out")
	if err != nil {
		t.Fatalf("os.Create()=%v", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("f.Close()=%v", err)
		}
	}()

	err = trace.Start(f)
	if err != nil {
		t.Fatalf("trace.Start()=%v", err)
	}
	defer trace.Stop()

	cycles := 1000
	var totalCycles atomic.Int64
	var totalSum atomic.Int64
	wg := sync.WaitGroup{}

	for i := range 4 {
		_ = i
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime.LockOSThread()
			var localSum int64
			for j := range cycles {
				_ = j
				for k := range 100000 {
					localSum += int64(k * rand.IntN(100))
				}
			}
			totalCycles.Add(int64(cycles))
			totalSum.Add(localSum)
		}()
	}

	wg.Wait()

	log.Printf("totalCycles=%d, totalSum=%d", totalCycles.Load(), totalSum.Load())
}

func TestTrace1(t *testing.T) {
	f, err := os.Create("trace1.out")
	if err != nil {
		t.Fatalf("os.Create()=%v", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			log.Printf("f.Close()=%v", err)
		}
	}()

	err = trace.Start(f)
	if err != nil {
		t.Fatalf("trace.Start()=%v", err)
	}
	defer trace.Stop()

	cycles := 1000
	var totalCycles atomic.Int64
	var totalSum atomic.Int64
	wg := sync.WaitGroup{}

	for i := range 28 {
		_ = i
		wg.Add(1)
		go func() {
			defer wg.Done()
			var localSum int64
			//runtime.LockOSThread()
			data := make([]int64, 4*65536)
			for idx := range data {
				data[idx] = int64(rand.IntN(100000))
			}
			for j := range cycles {
				_ = j
				for k := range data {
					localSum += int64(k) * data[k]
				}
			}
			totalCycles.Add(int64(cycles))
			totalSum.Add(localSum)
		}()
	}

	wg.Wait()

	log.Printf("totalCycles=%d, totalSum=%d", totalCycles.Load(), totalSum.Load())
}
