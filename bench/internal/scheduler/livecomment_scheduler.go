package scheduler

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

var LivecommentScheduler = mustNewLivecommentScheduler()

// スパムを取り出す (ただし、なるべく投稿数の少ないスパム)
// ライブコメントを取り出す (ただし、なるべく投稿数の少ないライブコメント)
// チップを取り出す
//// チップレベルを指定したら、それに合わせて金額を返すように

// ライブコメント数、スパム数などに応じて投げ銭するモチベーションを制御したい
// ただし、ゲーム性を損なわない範囲にしたいので、投げ銭してもらうまでの難易度が上がるというようにしたい

// 予約後、ライブ配信の処理が重くなるように、ライブコメント(+投げ銭)やリアクションなどを管理し、考える
// 投げ銭が偏るように采配するか、偏らないように分散させるか

// 配信の種類を決める
// * 通常
// * 人気
// * 炎上

// 炎上ノルマ達成か？
// 人気ノルマ達成か？
// などのメソッドをはやし、呼び出し側で未達成なら炎上配信者払い出しなどというふうにする
// 炎上配信者は、可能な限り人気があると良い
// 人気は、もちろん人気がまだないことが条件
// それ以外、通常に分類され、ユーザは通常配信者と視聴者になる

type Livecomment struct {
	UserID       int
	LivestreamID int
	Comment      string
	Tip          int
}

type PositiveComment struct {
	Comment string
}

type NegativeComment struct {
	Comment string
	NgWord  string
}

// どの配信に対して色々投げたらいいか、いい感じにしてくれる君

// Positiveの方は、長いコメント、短いコメントみたいな感じで取れると良い

// シナリオを書く際の疑問を列挙しよう
// どこにスパムを投げればいい？
// どこにスパチャを投げると、平等に投げられそう？
// 人気配信はどこ？そこにスパチャや投げ銭を集中させたい
//    人気配信は、人気ユーザに紐づく配信が用いられる
//

// ポジティブ？長い？といった、どういうコメントを取得するかは取得側で判断
//

type StreamerStatistics struct {
	NumLivecomments   int
	TotalTips         int
	TotalReportsCount int
}

type livecommentScheduler struct {
	// 配信者ごと、ライブコメント数、投げ銭売上合計、スパム数の統計を取る
	// 構造体は全部Livecommentなので、Commit, Abortを用意すればいいか
	// ライブコメント、投げ銭は投稿時でどちらも扱えるけど、スパムはスパムメッセージなのかスパム報告なのか難しいな
	streamerStats map[int]*StreamerStatistics
}

func mustNewLivecommentScheduler() *livecommentScheduler {
	return &livecommentScheduler{}
}

// FIXME:

func (s *livecommentScheduler) GetShortPositiveComment() *PositiveComment {
	idx := rand.Intn(len(positiveCommentPool))
	return positiveCommentPool[idx]
}

func (s *livecommentScheduler) GetLongPositiveComment() *PositiveComment {
	idx := rand.Intn(len(positiveCommentPool))
	return positiveCommentPool[idx]
}

func (s *livecommentScheduler) GetNegativeComment() *NegativeComment {
	idx := rand.Intn(len(negativeCommentPool))
	return negativeCommentPool[idx]
}

// 通常配信に対するチップ取得
func (s *livecommentScheduler) GetTipsForStream() int {
	return GenerateIntBetween(1, 4)
}

// 人気配信に対するチップ取得
func (s *livecommentScheduler) GetTipsForPopularStream() int {
	n := rand.Intn(2)
	if n == 1 {
		return 5
	} else {
		return s.GetTipsForStream()
	}
}
