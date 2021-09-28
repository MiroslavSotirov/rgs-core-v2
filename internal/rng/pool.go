package rng

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	generators []*rand.Rand
	available  []int32
	lock       sync.Mutex
}

func (pool *Pool) insert(rng *rand.Rand) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.generators = append(pool.generators, rng)
	pool.available = append(pool.available, 0)
}

func (pool *Pool) New() *rand.Rand {
	mt := NewRNG()
	b := make([]byte, 32)
	_, err := cryptorand.Read(b)
	if err != nil {
		panic("could not read entropy to seed mt19937 rng")
	}
	seed := make([]uint64, 4)
	for i := range seed {
		seed[i] = binary.LittleEndian.Uint64(b[:8])
		b = b[8:]
	}
	mt.SeedFromSlice(seed)
	rng := rand.New(mt)
	pool.insert(rng)
	return rng
}

func (pool *Pool) Get() *rand.Rand {
	for idx := range pool.available {
		if atomic.CompareAndSwapInt32(&pool.available[idx], 1, 0) {
			return pool.generators[idx]
		}
	}
	return pool.New()
}

func (pool *Pool) Put(rng *rand.Rand) {
	for idx, gen := range pool.generators {
		if gen == rng {
			pool.available[idx] = 1
			return
		}
	}
	panic("put with a rng that is not part of the pool")
}

func CyclePool(pool *Pool) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for true {
		for idx := range pool.generators {
			if atomic.CompareAndSwapInt32(&pool.available[idx], 1, 0) {
				_ = pool.generators[idx].Uint64()
				pool.available[idx] = 1
			}
		}
		time.Sleep(time.Duration(rng.Int63n(1e9 / 4))) // 0 - 250ms
	}
}
