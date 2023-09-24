package generator

import "math/rand"

// generateIntBetween generates integer satisfies [min, max) constraint
func generateIntBetween(min, max int) int {
	return rand.Intn(max-min) + min
}
