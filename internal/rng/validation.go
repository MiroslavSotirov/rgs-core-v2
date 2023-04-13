package rng

import (
	"os"
)

/*
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
*/
func PrintRandomBytes() {

	//	gen := rand.New(cryptoSource{}) // rand.NewSource(time.Now().UnixNano()))

	for true {
		buf := make([]byte, 1024)
		RandBytes(buf)
		// gen.Read(buf)
		os.Stdout.Write(buf)
	}
}
