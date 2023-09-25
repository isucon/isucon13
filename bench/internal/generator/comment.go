package generator

var randomSuperchatComments = []string{
	"こんいす！",
	"こんいすー",
	"こんいす〜",
	"いす！",
	"ないす！",
	// FIXME: スパム判定されるコメントも追加する
}

func GenerateRandomComment() string {
	return randomSuperchatComments[randomSource.Intn(len(randomSuperchatComments))]
}
