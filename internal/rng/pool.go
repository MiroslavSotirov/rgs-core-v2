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
	seeds      [][]uint64
}

func (pool *Pool) insert(rng *MT19937, seed []uint64) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.generators = append(pool.generators, rng)
	pool.available = append(pool.available, 1)
	/*
		for _, s := range pool.seeds {
			for i, v := range s {
				if v != seed[i] {
					break
				}
				if i == len(seed)-1 {
					panic("seed already in use")
				}
			}
		}
		pool.seeds = append(pool.seeds, seed)
	*/
	logger.Infof("Pooled mt19937 number %d", len(pool.generators))
}

func (pool *Pool) New() *MT19937 {
	rng := NewRNG()
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic("could not read entropy to seed mt19937 rng")
	}
	seed := make([]uint64, 4)
	for i := range seed {
		seed[i] = binary.LittleEndian.Uint64(b[:8])
		b = b[8:]
	}
	//	seed := int64(binary.LittleEndian.Uint64(b))
	//	rng.Seed(seed)
	rng.SeedFromSlice(seed)
	pool.insert(rng, seed)
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
