package rng

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
	"sync/atomic"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type Pool struct {
	generators []*MT19937
	available  []int32
	lock       sync.Mutex
}

func (pool *Pool) insert(rng *MT19937) {
	pool.lock.Lock()
	pool.generators = append(pool.generators, rng)
	pool.available = append(pool.available, 1)
	logger.Infof("Pooled mt19937 number %d", len(pool.generators))
	pool.lock.Unlock()
}

func (pool *Pool) New() *MT19937 {
	rng := NewRNG()
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic("could not read entropy to seed mt19937 rng")
	}
	seed := int64(binary.LittleEndian.Uint64(b))
	rng.Seed(seed)
	pool.insert(rng)
	return rng
}

func (pool *Pool) Get() *MT19937 {
	for idx := range pool.available {
		if atomic.CompareAndSwapInt32(&pool.available[idx], 1, 0) {
			return pool.generators[idx]
		}
	}
	return pool.New()
}

func (pool *Pool) Put(rng *MT19937) {
	for idx, gen := range pool.generators {
		if gen == rng {
			pool.available[idx] = 1
			return
		}
	}
	pool.insert(rng)
}
