package generator

var randomReactionEmojiNames = []string{
	":innocent:",
	":tada:",
	":chair:",
	":+1:",
}

func GenerateRandomReaction() string {
	return randomReactionEmojiNames[randomSource.Intn(len(randomReactionEmojiNames))]
}
