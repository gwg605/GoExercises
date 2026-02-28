package keyvaluestore_test

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
	"valerygordeev/go/exercises/keyvaluestore"
)

func TestKVChannel(t *testing.T) {
	workersCount := 4
	keysCount := 1000
	data := make([]byte, 4096)

	for w := range workersCount {
		kv := keyvaluestore.NewKVChannels(w + 1)

		for i := range keysCount {
			key := fmt.Sprintf("key[%d]", i)
			err := kv.Set(key, data, time.Hour)
			if err != nil {
				t.Fatalf("TestKVChannel[%d]::Set[%d] failed. Error=%v", w, i, err)
			}
		}

		for i := range keysCount {
			key := fmt.Sprintf("key[%d]", i)
			data, err := kv.Get(key)
			if err != nil {
				t.Fatalf("TestKVChannel[%d]::Get[%d] failed. Error=%v", w, i, err)
			}
			if data == nil {
				t.Fatalf("TestKVChannel[%d]::Get[%d] no data", w, i)
			}
		}

		kv.Close()
	}

	log.Printf("Done!")
}

func client(w int, kv keyvaluestore.KVService, keysCount int) {
	data := make([]byte, 4096)
	//setStartTime := time.Now()

	for i := range keysCount {
		key := fmt.Sprintf("key[%d]", i)
		err := kv.Set(key, data, time.Hour)
		if err != nil {
			log.Printf("TestKVChannel[%d]::Set[%d] failed. Error=%v", w, i, err)
		}
	}

	//log.Printf("client[%d] - set: count=%d, time=%v", w, keysCount, time.Since(setStartTime))

	//getStartTime := time.Now()
	for i := range keysCount {
		key := fmt.Sprintf("key[%d]", i)
		data, err := kv.Get(key)
		if err != nil {
			log.Printf("TestKVChannel[%d]::Get[%d] failed. Error=%v", w, i, err)
		}
		if data == nil {
			log.Printf("TestKVChannel[%d]::Get[%d] no data", w, i)
		}
	}

	//log.Printf("client[%d] - get: count=%d, time=%v", w, keysCount, time.Since(getStartTime))
}

var kvChannels = keyvaluestore.NewKVChannels(64)
var kvMutexes = keyvaluestore.NewKVMutexes(64)

func BenchmarkKVChannel(b *testing.B) {
	wg := sync.WaitGroup{}

	clientsCount := 1024
	wg.Add(clientsCount)
	for i := range clientsCount {
		go func() {
			client(i, kvChannels, max(b.N/clientsCount, 1))
			wg.Done()
		}()
	}
	wg.Wait()
}

func BenchmarkKVMutexes(b *testing.B) {
	wg := sync.WaitGroup{}

	clientsCount := 1024
	wg.Add(clientsCount)
	for i := range clientsCount {
		go func() {
			client(i, kvMutexes, max(b.N/clientsCount, 1))
			wg.Done()
		}()
	}
	wg.Wait()
}
