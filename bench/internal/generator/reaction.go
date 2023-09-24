package generator

import "math/rand"

var randomReactionEmojiNames = []string{
	":innocent:",
	":tada:",
	":chair:",
	":+1:",
}

type RandomReactionGenerator struct {
}

func NewRandomReactionGenerator() *RandomReactionGenerator {
	return &RandomReactionGenerator{}
}

func (r *RandomReactionGenerator) Generate() string {
	return randomReactionEmojiNames[rand.Intn(len(randomReactionEmojiNames))]
}
