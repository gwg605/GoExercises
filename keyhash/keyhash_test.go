package keyhash_test

import (
	"fmt"
	"iter"
	"log"
	"math"
	"testing"
	"valerygordeev/go/exercises/common"
	"valerygordeev/go/exercises/keyhash"

	"github.com/google/uuid"
)

type Hasher func(string) uint64

type KeysIterator iter.Seq[string]

type ShardKeysStats []int

func CollectShardKeysStats(hasher Hasher, shards int, keys []string) ShardKeysStats {
	result := make(ShardKeysStats, shards)
	for _, key := range keys {
		hash := hasher(key)
		shard := byte(hash) % byte(shards)
		result[shard]++
	}
	return result
}

func DumpShardKeysStat(label string, stat ShardKeysStats) {
	var totalKeys int
	for _, keysCount := range stat {
		totalKeys += keysCount
	}
	shardsCount := max(len(stat), 1)
	averageKeysCount := totalKeys / shardsCount
	var maxDisp = math.MinInt
	var minDisp = math.MaxInt
	var totalDisp = 0
	for _, keysCount := range stat {
		disp := common.AbsInt(keysCount - averageKeysCount)
		if disp < minDisp {
			minDisp = disp
		}
		if disp > maxDisp {
			maxDisp = disp
		}
		totalDisp += disp
	}
	if minDisp == math.MaxInt {
		minDisp = 0
		maxDisp = 0
	}
	averageDisp := totalDisp / shardsCount
	log.Printf("%s|%9d|%7.03f|%9d|%7.03f|%9d|%7.03f|%9d| %v",
		label, averageKeysCount,
		float64(minDisp)/float64(averageKeysCount), minDisp,
		float64(averageDisp)/float64(averageKeysCount), averageDisp,
		float64(maxDisp)/float64(averageKeysCount), maxDisp,
		stat)
}

type KeysIteratorGetter func(int) KeysIterator

func GetBasicKeysSeq(keysCount int) KeysIterator {
	return func(yield func(string) bool) {
		for i := range keysCount {
			key := fmt.Sprintf("key[%d]", i)
			if !yield(key) {
				return
			}
		}
	}
}

func GetUUIDKeysSeq(keysCount int) KeysIterator {
	return func(yield func(string) bool) {
		for i := range keysCount {
			_ = i
			id, err := uuid.NewUUID()
			if err != nil {
				log.Printf("uuid.NewUUID() failed. Error=%v", err)
				return
			}
			if !yield(id.String()) {
				return
			}
		}
	}
}

func GetUUIDv7KeysSeq(keysCount int) KeysIterator {
	return func(yield func(string) bool) {
		for i := range keysCount {
			_ = i
			id, err := uuid.NewV7()
			if err != nil {
				log.Printf("uuid.NewV7() failed. Error=%v", err)
				return
			}
			if !yield(id.String()) {
				return
			}
		}
	}
}

func TestKeyHashers(t *testing.T) {
	keysCounts := []int{100000, 500000, 1000000}
	keysCountsMax := keysCounts[len(keysCounts)-1]
	hashers := map[string]Hasher{
		"KeyHash": keyhash.KeyHash,
		"HashFnv": keyhash.HashFnv,
		"hashMap": keyhash.HashMap,
	}
	keysPatterms := map[string]KeysIteratorGetter{
		"key[<index>]": GetBasicKeysSeq,
		"<uuid>":       GetUUIDKeysSeq,
		"<uuid.v7>":    GetUUIDv7KeysSeq,
	}

	keysSeqs := map[string][]string{}
	for keysPattern, keysSeqGetter := range keysPatterms {
		keys := make([]string, 0, keysCountsMax)
		for v := range keysSeqGetter(keysCountsMax) {
			keys = append(keys, v)
		}
		keysSeqs[keysPattern] = keys
	}

	log.Printf("Hasher | Keys pattern |  Keys  |Shards|KeysShard| Min, %%|Min, unit| Avg, %%|Avg, unit| Max, %%|Max, unit| Keys in Shards")
	log.Printf("-------+--------------+--------+------+---------+--------+---------+--------+---------+--------+---------+---------------------------------------------------------")
	for keysPattern := range keysPatterms {
		for shardsCount := 8; shardsCount <= 16; shardsCount *= 2 {
			for _, keysCount := range keysCounts {
				for hasherName, hasher := range hashers {
					stat := CollectShardKeysStats(hasher, shardsCount, keysSeqs[keysPattern][:keysCount])
					label := fmt.Sprintf("%s|%14s|%8d|%6d", hasherName, keysPattern, keysCount, shardsCount)
					DumpShardKeysStat(label, stat)
				}
			}
		}
	}
}

func BenchmarkKeyHashBasicSeq(b *testing.B) {
	for v := range GetBasicKeysSeq(1) {
		for i := range b.N {
			_ = i
			_ = keyhash.KeyHash(v)
		}
	}
}

func BenchmarkHashFnvBasicSeq(b *testing.B) {
	for v := range GetBasicKeysSeq(1) {
		for i := range b.N {
			_ = i
			_ = keyhash.HashFnv(v)
		}
	}
}

func BenchmarkHashMapBasicSeq(b *testing.B) {
	for v := range GetBasicKeysSeq(1) {
		for i := range b.N {
			_ = i
			_ = keyhash.HashMap(v)
		}
	}
}

func BenchmarkKeyHashUUIDSeq(b *testing.B) {
	for v := range GetUUIDKeysSeq(1) {
		for i := range b.N {
			_ = i
			_ = keyhash.KeyHash(v)
		}
	}
}

func BenchmarkHashFnvUUIDSeq(b *testing.B) {
	for v := range GetUUIDKeysSeq(1) {
		for i := range b.N {
			_ = i
			_ = keyhash.HashFnv(v)
		}
	}
}

func BenchmarkHashMapUUIDSeq(b *testing.B) {
	for v := range GetUUIDKeysSeq(1) {
		for i := range b.N {
			_ = i
			_ = keyhash.HashMap(v)
		}
	}
}
