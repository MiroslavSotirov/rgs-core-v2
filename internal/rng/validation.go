package rng

import (
	"os"
)

func PrintRandomBytes() {

	for true {
		buf := make([]byte, 1024)
		RandBytes(buf)
		os.Stdout.Write(buf)
	}
}
