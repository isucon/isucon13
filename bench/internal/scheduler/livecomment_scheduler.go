package scheduler

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/isucon/isucon13/bench/internal/bencherror"
)

var randomSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// GenerateIntBetween generates integer satisfies [min, max) constraint
func generateTipValueBetween(min, max int) int {
	v := randomSource.Intn(max-min) + min

	// 桁数
	var numDigit int
	for target := v; target > 0; {
		target = target / 10
		numDigit++
	}

	if v >= 10 && numDigit >= 2 {
		var (
			quotient = int(math.Pow(10, float64(numDigit)-1))
			surplus  = v % quotient
		)
		return v - surplus
	} else {
		return v
	}
}

var LivecommentScheduler = mustNewLivecommentScheduler()

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

type NgWord struct {
	Word string
}

type Tip struct {
	Level int
	Tip   int
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
type livecommentScheduler struct {
	ngLivecomments map[string]string

	moderatedMu sync.RWMutex
	moderated   map[string]struct{}
}

func mustNewLivecommentScheduler() *livecommentScheduler {
	ngLivecomments := make(map[string]string)
	for _, comment := range negativeCommentPool {
		ngLivecomments[comment.Comment] = comment.NgWord
	}
	rand.Shuffle(len(dummyNgWords), func(i, j int) {
		dummyNgWords[i], dummyNgWords[j] = dummyNgWords[j], dummyNgWords[i]
	})
	return &livecommentScheduler{
		ngLivecomments: ngLivecomments,
		moderated:      make(map[string]struct{}),
	}
}

// ライブコメント一覧に何件スパムが含まれるか調べるために使う
func (s *livecommentScheduler) IsNgLivecomment(comment string) bool {
	if _, ok := s.ngLivecomments[comment]; ok {
		return true
	} else {
		return false
	}
}

func (s *livecommentScheduler) GetNgWord(comment string) (string, error) {
	ngword, ok := s.ngLivecomments[comment]
	if !ok {
		return "", bencherror.NewInternalError(fmt.Errorf("想定されているスパムコメントではありません: %s", comment))
	}

	return ngword, nil
}

func (s *livecommentScheduler) GetShortPositiveComment() *PositiveComment {
	idx := rand.Intn(len(positiveCommentPool))
	return positiveCommentPool[idx]
}

func (s *livecommentScheduler) GetLongPositiveComment() *PositiveComment {
	idx := rand.Intn(len(positiveCommentPool))
	return positiveCommentPool[idx]
}

func (s *livecommentScheduler) GetNegativeComment() (*NegativeComment, bool) {
	s.moderatedMu.RLock()
	defer s.moderatedMu.RUnlock()

	idx := rand.Intn(len(negativeCommentPool))
	comment := negativeCommentPool[idx]
	_, isModerated := s.moderated[comment.Comment]
	return comment, isModerated
}

func (s *livecommentScheduler) IsModerated(comment string) bool {
	s.moderatedMu.RLock()
	defer s.moderatedMu.RUnlock()

	_, isModerated := s.moderated[comment]
	return isModerated
}

func (s *livecommentScheduler) Moderate(comment string) {
	s.moderatedMu.Lock()
	defer s.moderatedMu.Unlock()

	s.moderated[comment] = struct{}{}
}

func (s *livecommentScheduler) generateTip(level int) int {
	switch level {
	case 0:
		return 0
	case 1:
		return generateTipValueBetween(10, 100)
	case 2:
		return generateTipValueBetween(100, 1000)
	case 3:
		return generateTipValueBetween(1000, 5000)
	case 4:
		return generateTipValueBetween(5000, 10000)
	case 5:
		return generateTipValueBetween(10000, 100000)
	default:
		return 0
	}
}

// 通常配信に対するチップ取得
func (s *livecommentScheduler) GetTipsForStream(totalHours, currentHour int) *Tip {
	if currentHour > totalHours {
		log.Fatalf("GetTipsForPopularStream: 引数が不正です currentHour=%d > totalHours=%d\n", currentHour, totalHours)
		return &Tip{Level: 0, Tip: 0}
	}
	if totalHours == 1 {
		return &Tip{
			Level: 1,
			Tip:   s.generateTip(1),
		}
	}
	var level int
	threshold := totalHours / 2
	if currentHour >= threshold {
		level = 2
	} else {
		level = 1
	}

	tip := s.generateTip(level)
	return &Tip{
		Level: level,
		Tip:   tip,
	}
}

// 長時間配信に対するチップ取得
// 10h以上を想定
func (s *livecommentScheduler) GetTipsForLongStream(totalHours, currentHour int) *Tip {
	if totalHours < 10 {
		log.Fatalf("GetTipsForPopularStream: 時間枠が短すぎます: totalHours=%d\n", totalHours)
		return &Tip{Level: 0, Tip: 0}
	}
	if currentHour > totalHours {
		log.Fatalf("GetTipsForPopularStream: 引数が不正です currentHour=%d > totalHours=%d\n", currentHour, totalHours)
		return &Tip{Level: 0, Tip: 0}
	}
	if totalHours < 1 {
		return &Tip{
			Level: 0,
			Tip:   0,
		}
	}
	threshold := totalHours / 5
	level := min(1+(currentHour/threshold), 5)

	tip := s.generateTip(level)
	return &Tip{
		Level: level,
		Tip:   tip,
	}
}

// 3 6 9 12 15
// thres=3
// 1 2 3 4 5

func (s *livecommentScheduler) GetDummyNgWord() *NgWord {
	idx := rand.Intn(len(dummyNgWords))
	return dummyNgWords[idx]
}
