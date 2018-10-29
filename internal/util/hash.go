package util

import "github.com/spaolacci/murmur3"

func Hash(key []byte) uint64 {
	hasher := murmur3.New64()
	hasher.Write(key)
	return hasher.Sum64()
}
