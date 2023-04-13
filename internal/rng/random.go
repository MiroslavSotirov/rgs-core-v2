package rng

import (
	"encoding/base64"
	"math/rand"
	"strings"

	uuid "github.com/satori/go.uuid"
)

var rngPool = Pool{}

func Init() {
	rngPool.Put(rngPool.Get())
	go CyclePool(&rngPool)
}

// RandStringRunes ...
func RandStringRunes(n int) string {
	rng := rngPool.Get()
	defer rngPool.Put(rng)
	return randStringRunes(rng, n)
}

// RandFromRange
func RandFromRange(n int) int {
	rng := rngPool.Get()
	defer rngPool.Put(rng)
	return randFromRange(rng, n)
}

func RandBytes(b []byte) {
	rng := rngPool.Get()
	defer rngPool.Put(rng)
	randBytes(rng, b)
}

func Uuid() string {
	return strings.ReplaceAll(base64.StdEncoding.EncodeToString(uuid.NewV4().Bytes()), "/", "_")
}

func randStringRunes(rng *rand.Rand, n int) string {
	b := make([]byte, n-n/4)
	rng.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func randFromRange(rng *rand.Rand, n int) int {
	// returns a random integer from 0 to n-1
	return int(rng.Uint64() % uint64(n))
}

func randBytes(rng *rand.Rand, b []byte) {
	rng.Read(b)
}
