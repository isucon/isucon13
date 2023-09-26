package generator

import (
	"math/rand"
	"time"
)

var (
	randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// GenerateIntBetween generates integer satisfies [min, max) constraint
func GenerateIntBetween(min, max int) int {
	return randomSource.Intn(max-min) + min
}
