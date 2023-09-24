package generator

import (
	"log"
)

type TipLevel int

// FIXME: スコアと絡めた考慮
const (
	TipLevel1 TipLevel = iota
	TipLevel2
	TipLevel3
	TipLevel4
	TipLevel5
)

// GenerateTip は、投げ銭を生成します
// チップレベルは、チップとスコアの関係を表現します。あるチップが与えられたとき、ベンチマークスコアが何点になるかを決定します
// FIXME: tipは１つのパッケージにまとめたほうがいいかも
func GenerateTip(level TipLevel) int {
	switch level {
	case TipLevel1:
		return generateIntBetween(10, 100)
	case TipLevel2:
		return generateIntBetween(100, 200)
	case TipLevel3:
		return generateIntBetween(200, 300)
	case TipLevel4:
		return generateIntBetween(300, 400)
	case TipLevel5:
		return generateIntBetween(400, 500)
	default:
		// panic & non-return
		log.Fatalf("uncovered tip level specified: %d\n", level)
		return 0
	}
}
