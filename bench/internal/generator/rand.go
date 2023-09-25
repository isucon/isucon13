package generator

import (
	"math/rand"
	"time"
)

var (
	randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// generateIntBetween generates integer satisfies [min, max) constraint
func generateIntBetween(min, max int) int {
	return randomSource.Intn(max-min) + min
}
