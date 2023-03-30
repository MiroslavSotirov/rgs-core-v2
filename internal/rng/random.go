package rng

import (
	"encoding/base64"
	"math/rand"
	"strings"

	uuid "github.com/satori/go.uuid"
)

var rngPool = Pool{}

func InitPool() {
	rngPool.Put(rngPool.Get())
	go CyclePool(&rngPool)
}

// RandStringRunesPool ...
func RandStringRunesPool(n int) string {
	rng := rngPool.Get()
	defer rngPool.Put(rng)
	return randStringRunesPool(rng, n)
}

// RandFromRangePool
func RandFromRangePool(n int) int {
	rng := rngPool.Get()
	defer rngPool.Put(rng)
	return randFromRangePool(rng, n)
}

func Uuid() string {
	return strings.ReplaceAll(base64.StdEncoding.EncodeToString(uuid.NewV4().Bytes()), "/", "_")
}

func randStringRunesPool(rng *rand.Rand, n int) string {
	b := make([]byte, n-n/4)
	rng.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

/*
func (rng *MT19937) RandStringRunesPool(n int) string {
	t := time.Now()
	b := make([]rune, n)

	for i := range b {
		randomIndex := int(rng.Uint64() % uint64(float32(len(runes))))
		b[i] = runes[randomIndex]
	}
	d := time.Now().Sub(t)
	logger.Infof("RandStringRunesPool: \"%s\" with length %d in %.4fms", string(b), len(string(b)), float64(d)/1000000.0)
	return string(b)
}
*/

func randFromRangePool(rng *rand.Rand, n int) int {
	// returns a random integer from 0 to n-1
	return int(rng.Uint64() % uint64(n))
}
