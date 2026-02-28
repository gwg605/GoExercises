package keyvaluestore

import (
	"fmt"
	"sync"
	"time"
	"valerygordeev/go/exercises/keyhash"
)

type KVMutexes struct {
	locks  []sync.RWMutex
	stores []map[string]CacheRecord
}

func NewKVMutexes(shards int) *KVMutexes {
	result := &KVMutexes{
		locks:  make([]sync.RWMutex, shards),
		stores: make([]map[string]CacheRecord, shards),
	}
	for i := range shards {
		result.locks[i] = sync.RWMutex{}
		result.stores[i] = make(map[string]CacheRecord)
	}
	return result
}

func (kv *KVMutexes) Set(key string, data []byte, ttl time.Duration) error {
	keyHash := keyhash.KeyHash(key)
	storeIndex := byte(keyHash) % byte(len(kv.stores))

	kv.locks[storeIndex].Lock()
	defer kv.locks[storeIndex].Unlock()

	kv.stores[storeIndex][key] = CacheRecord{Data: data, Expire: time.Now().UTC().Add(ttl)}
	return nil
}

func (kv *KVMutexes) Get(key string) ([]byte, error) {
	keyHash := keyhash.KeyHash(key)
	storeIndex := byte(keyHash) % byte(len(kv.stores))

	kv.locks[storeIndex].RLock()
	defer kv.locks[storeIndex].RUnlock()

	record, found := kv.stores[storeIndex][key]
	if found {
		if record.Expire.After(time.Now().UTC()) {
			return record.Data, nil
		}
	}
	return nil, fmt.Errorf("No key")
}

func (kv *KVMutexes) Close() {
}
