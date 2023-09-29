package generator

var randomLivecommentComments = []string{
	"こんいす！",
	"こんいすー",
	"こんいす〜",
	"いす！",
	"ないす！",
	// FIXME: スパム判定されるコメントも追加する
}

func GenerateRandomComment() string {
	return randomLivecommentComments[randomSource.Intn(len(randomLivecommentComments))]
}
