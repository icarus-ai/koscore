package crypto

import (
	mrand "math/rand"
	mrand2 "math/rand/v2"

	"crypto/rand"
	"time"
)

func RandomBytes(size int) []byte {
	b := make([]byte, size)
	_, _ = rand.Read(b)
	return b
}

func RandomI6(min, max int64) int64 {
	return min + mrand.New(mrand.NewSource(time.Now().UnixNano())).Int63n(max-min)
}

var RandU32 = mrand2.Uint32

// [0,max-min] + min => [min,max]
func RandomU32(min, max uint32) uint32 { return (mrand2.Uint32() & (max - min)) + min }
func RandomI64(min, max int64) int64   { return (mrand2.Int64() & (max - min)) + min }
