package scheduler

import (
	"fmt"
	"math"
	"math/rand"
	"sync"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/config"
)

// 数値の端数を落とす
func trimFraction(v int) int {
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

func (s *livecommentScheduler) ModerateNgWord(ngword string) {
	s.moderatedMu.Lock()
	defer s.moderatedMu.Unlock()

	for _, comment := range negativeCommentPool {
		if comment.NgWord == ngword {
			s.moderated[comment.Comment] = struct{}{}
		}
	}
}

func (s *livecommentScheduler) generateTip(level int, totalHours, currentHour int) int {
	progressRate := currentHour / totalHours
	switch level {
	case 0:
		return 0
	case 1:
		var (
			minTip = 10
			maxTip = 100
		)
		return ((maxTip - minTip) * progressRate) + minTip
	case 2:
		var (
			minTip = 100
			maxTip = 1000
		)
		return ((maxTip - minTip) * progressRate) + minTip
	case 3:
		var (
			minTip = 1000
			maxTip = 5000
		)
		return ((maxTip - minTip) * progressRate) + minTip
	case 4:
		var (
			minTip = 5000
			maxTip = 10000
		)
		return ((maxTip - minTip) * progressRate) + minTip
	case 5:
		var (
			minTip = 10000
			maxTip = 100000
		)
		return ((maxTip - minTip) * progressRate) + minTip
	default:
		return 0
	}
}

func (s *livecommentScheduler) GetTipsForStream(totalHours, currentHour int) (*Tip, error) {
	if currentHour > totalHours {
		return &Tip{Level: 0, Tip: 0}, bencherror.NewInternalError(fmt.Errorf("GetTipsForStreamの引数が不正です: current=%d, total=%d", currentHour, totalHours))
	}
	if totalHours < 1 || currentHour < 1 {
		return &Tip{Level: 0, Tip: 0}, bencherror.NewInternalError(fmt.Errorf("GetTipsForStreamの引数が不正です: current=%d, total=%d", currentHour, totalHours))
	}

	// levelによって金額クラスが分かれる. より長い配信枠のほうが高いレベルになる. 予約が捌けているほど高いレベルになる.
	// level内ではどれだけ視聴し続けられたかが評価される. これも長い配信枠のほうがよりTipが高額になるが、それだけでなくwebappがライブコメント投稿を捌けていないと高額にならない
	var level int
	switch {
	case totalHours >= 20:
		level = 5
	case totalHours >= 15:
		level = 4
	case totalHours >= config.LongHourThreshold:
		level = 3
	case totalHours >= 5:
		level = 2
	default:
		level = 1
	}

	tip := s.generateTip(1, totalHours, currentHour)
	return &Tip{
		Level: level,
		Tip:   tip,
	}, nil
}

func (s *livecommentScheduler) GetDummyNgWord() *NgWord {
	idx := rand.Intn(len(dummyNgWords))
	return dummyNgWords[idx]
}
