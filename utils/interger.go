package utils

import (
	"math/rand"
	"time"
)

func RandRangeUint32(min, max int32) uint32 {
	if min >= max || min == 0 || max == 0 {
		return uint32(max)
	}
	rand.Seed(time.Now().UnixNano())
	return uint32(rand.Int31n(max-min) + min)
}
