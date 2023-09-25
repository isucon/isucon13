package generator

import "math/rand"

var randomReactionEmojiNames = []string{
	":innocent:",
	":tada:",
	":chair:",
	":+1:",
}

func GenerateRandomReaction() string {
	return randomReactionEmojiNames[rand.Intn(len(randomReactionEmojiNames))]
}
