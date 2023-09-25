package generator

import "math/rand"

var randomSuperchatComments = []string{
	"こんいす！",
	"こんいすー",
	"こんいす〜",
	"いす！",
	"ないす！",
	// FIXME: スパム判定されるコメントも追加する
}

func GenerateRandomComment() string {
	return randomSuperchatComments[rand.Intn(len(randomSuperchatComments))]
}
