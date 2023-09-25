package generator

import (
	"log"
)

type TipLevel int

// FIXME: スコアと絡めた考慮
const (
	TipLevel0 TipLevel = iota
	TipLevel1
	TipLevel2
	TipLevel3
	TipLevel4
	TipLevel5
)

var tipLevels = []TipLevel{
	TipLevel0,
	TipLevel1,
	TipLevel2,
	TipLevel3,
	TipLevel4,
	TipLevel5,
}

func GenerateRandomTipLevel() TipLevel {
	return tipLevels[(randomSource.Intn(len(tipLevels)))]
}

// GenerateTip は、投げ銭を生成します
// チップレベルは、チップとスコアの関係を表現します。あるチップが与えられたとき、ベンチマークスコアが何点になるかを決定します
func GenerateTip(level TipLevel) int {
	switch level {
	case TipLevel0:
		return 0
	case TipLevel1:
		return generateIntBetween(1, 500)
	case TipLevel2:
		return generateIntBetween(500, 1000)
	case TipLevel3:
		return generateIntBetween(1000, 5000)
	case TipLevel4:
		return generateIntBetween(5000, 10000)
	case TipLevel5:
		return generateIntBetween(10000, 20001) // 半開区間において、20kを含めるため
	default:
		// panic & non-return
		log.Printf("uncovered tip level specified: %d\n", level)
		return 0
	}
}
