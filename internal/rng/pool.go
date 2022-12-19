package rng

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"log"
	mrand "math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type Pool struct {
	generators []*mrand.Rand
	available  []int32
	lock       sync.Mutex
}

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(cryptorand.Reader, binary.BigEndian, &v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func (pool *Pool) insert(rng *mrand.Rand) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	pool.generators = append(pool.generators, rng)
	pool.available = append(pool.available, 0)
}

func (pool *Pool) New() *mrand.Rand {
	src := &cryptoSource{}
	rng := mrand.New(src)
	pool.insert(rng)
	return rng
}

func (pool *Pool) Get() *mrand.Rand {
	for idx := range pool.available {
		if atomic.CompareAndSwapInt32(&pool.available[idx], 1, 0) {
			return pool.generators[idx]
		}
	}
	return pool.New()
}

func (pool *Pool) Put(rng *mrand.Rand) {
	for idx, gen := range pool.generators {
		if gen == rng {
			pool.available[idx] = 1
			return
		}
	}
	panic("put with a rng that is not part of the pool")
}

func CyclePool(pool *Pool) {
	var src cryptoSource
	rng := mrand.New(src)

	for {
		for idx := range pool.generators {
			if atomic.CompareAndSwapInt32(&pool.available[idx], 1, 0) {
				_ = pool.generators[idx].Uint64()
				pool.available[idx] = 1
			}
		}
		time.Sleep(time.Duration(rng.Int63n(1e9 / 4))) // 0 - 250ms
	}
}
