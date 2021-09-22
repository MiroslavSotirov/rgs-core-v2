package rng

import (
	"encoding/base64"
)

var rngPool = Pool{}

func Init() {
	rngPool.Put(rngPool.Get())
}

var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// RandStringRunes ...
func RandStringRunes(n int) string {
	rng := rngPool.Get()
	defer rngPool.Put(rng)
	return rng.randStringRunes(n)
}

// RandFromRange
func RandFromRange(n int) int {
	rng := rngPool.Get()
	defer rngPool.Put(rng)
	return rng.randFromRange(n)
}

func (rng *MT19937) randStringRunes(n int) string {
	b := make([]byte, n-n/4)
	rng.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

/*
func (rng *MT19937) randStringRunes(n int) string {
	t := time.Now()
	b := make([]rune, n)

	for i := range b {
		randomIndex := int(rng.Uint64() % uint64(float32(len(runes))))
		b[i] = runes[randomIndex]
	}
	d := time.Now().Sub(t)
	logger.Infof("randStringRunes: \"%s\" with length %d in %.4fms", string(b), len(string(b)), float64(d)/1000000.0)
	return string(b)
}
*/

func (rng *MT19937) randFromRange(n int) int {
	// returns a random integer from 0 to n-1
	return int(rng.Uint64() % uint64(n))
}
