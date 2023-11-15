package benchscore

import (
	"sync/atomic"
)

var profit uint64

func AddTip(tip uint64) {
	atomic.AddUint64(&profit, tip)
}

// GetFinalProfit は、最終売上を返します
// FIXME: finalcheck後にprofitをスコアに加算しないと駄目
func GetTotalProfit() uint64 {
	return profit
}
