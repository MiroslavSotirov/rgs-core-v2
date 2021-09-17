package rng

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
	b := make([]rune, n)

	for i := range b {
		randomIndex := int(rng.Uint64() % uint64(float32(len(runes))))
		b[i] = runes[randomIndex]
	}
	return string(b)
}

func (rng *MT19937) randFromRange(n int) int {
	// returns a random integer from 0 to n-1
	return int(rng.Uint64() % uint64(n))
}
