package keyvaluestore

import (
	"fmt"
	"sync"
	"time"
	"valerygordeev/go/exercises/keyhash"
)

type MsgType int

const (
	MsgTypeNop MsgType = iota
	MsgTypeSet
	MsgTypeGet
)

type Msg struct {
	Type     MsgType
	Key      string
	Data     []byte
	TTL      time.Duration
	Complete chan struct{}
}

type KVChannels struct {
	channels []chan *Msg
	stores   []map[string]CacheRecord
	wg       sync.WaitGroup
}

func NewKVChannels(shards int) *KVChannels {
	result := &KVChannels{
		channels: make([]chan *Msg, shards),
		stores:   make([]map[string]CacheRecord, shards),
		wg:       sync.WaitGroup{},
	}
	for i := range shards {
		result.channels[i] = make(chan *Msg)
		result.stores[i] = make(map[string]CacheRecord)
		result.wg.Add(1)
		go result.workerRun(i)
	}
	return result
}

func (kv *KVChannels) workerRun(wi int) {
	//log.Printf("workerRun[%d] - start", wi)
	defer kv.wg.Done()
	for {
		msg := <-kv.channels[wi]
		if msg == nil {
			break
		}
		switch msg.Type {
		case MsgTypeSet:
			kv.stores[wi][msg.Key] = CacheRecord{Expire: time.Now().UTC().Add(msg.TTL), Data: msg.Data}
			msg.Complete <- struct{}{}
		case MsgTypeGet:
			record, found := kv.stores[wi][msg.Key]
			if found {
				if record.Expire.After(time.Now().UTC()) {
					msg.Data = record.Data
				}
			}
			msg.Complete <- struct{}{}
		}
	}
	//log.Printf("workerRun[%d] - finished", wi)
}

func (kv *KVChannels) Set(key string, data []byte, ttl time.Duration) error {
	keyHash := keyhash.KeyHash(key)
	storeIndex := byte(keyHash) % byte(len(kv.stores))
	msg := &Msg{Type: MsgTypeSet, Key: key, TTL: ttl, Data: data, Complete: make(chan struct{})}
	kv.channels[storeIndex] <- msg
	<-msg.Complete
	return nil
}

func (kv *KVChannels) Get(key string) ([]byte, error) {
	keyHash := keyhash.KeyHash(key)
	storeIndex := byte(keyHash) % byte(len(kv.stores))
	msg := &Msg{Type: MsgTypeGet, Key: key, Complete: make(chan struct{})}
	kv.channels[storeIndex] <- msg
	<-msg.Complete
	if msg.Data == nil {
		return nil, fmt.Errorf("No key")
	}
	return msg.Data, nil
}

func (kv *KVChannels) Close() {
	for _, channel := range kv.channels {
		close(channel)
	}
	kv.wg.Wait()
	//log.Printf("Closed!")
}
