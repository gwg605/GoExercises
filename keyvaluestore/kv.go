package keyvaluestore

import "time"

type CacheRecord struct {
	Expire time.Time
	Data   []byte
}

type KVService interface {
	Set(key string, val []byte, ttl time.Duration) error
	Get(key string) ([]byte, error)
	Close()
}
