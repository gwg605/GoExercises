package main

import (
	"hash/fnv"
	"hash/maphash"
)

var (
	maphashSeed = maphash.MakeSeed()
)

func KeyHash(key string) uint64 {
	var result = uint64(14695981039346656037)
	for _, b := range []byte(key) {
		result ^= uint64(b)
		result *= 1099511628211
	}
	return result
}

func HashMap(key string) uint64 {
	return maphash.String(maphashSeed, key)
}

func HashFnv(key string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(key))
	return h.Sum64()
}
