package scenario

import (
	"context"
	"log"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/isupipe"
)

const (
	// Season1GeneratedLivestreamCount は/initializeで注入されるseason1期間の配信レコード数
	Season1GeneratedLivestreamCount = 570
)

type loginUser struct {
	UserName string
	Password string
}

// userId -> user
var loginUsers = map[int]loginUser{
	1: {
		UserName: "井上 太郎",
		Password: "o^E0K1Axj@",
	},
	2: {
		UserName: "山崎 洋介",
		Password: "u4JVlvx%(6",
	},
	3: {
		UserName: "高橋 智也",
		Password: "Ba)6J7pmZY",
	},
	4: {
		UserName: "三浦 浩",
		Password: "@4$rPveY4b",
	},
	5: {
		UserName: "田中 洋介",
		Password: "m$C2hSyMac",
	},
	6: {
		UserName: "山田 裕美子",
		Password: "!4QdG!Ni&x",
	},
	7: {
		UserName: "山下 晃",
		Password: "+L_3kLjI61",
	},
	8: {
		UserName: "佐々木 さゆり",
		Password: ")i_7Qvnh1!",
	},
	9: {
		UserName: "高橋 太郎",
		Password: ")6ZVY&D1&v",
	},
	10: {
		UserName: "井上 春香",
		Password: ")22R(a&z%2",
	},
}

// Season1 シナリオは、サービス開始時点で存在する配信者の配信に対して、ランダムにリクエストを送信する
func Season1(ctx context.Context) error {
	log.Println("running season1 scenario ...")

	client, err := isupipe.NewClient(
		agent.WithBaseURL(config.TargetBaseURL),
	)
	if err != nil {
		return err
	}

	// 広告費用で制御して、リクエスト送信goroutineを単純倍増
	// INFO: リクエスト数を制御するだけでなく、tipsの金額も増加させても良いかもしれない
	for userIdx := 0; userIdx < config.AdvertiseCost; userIdx++ {
		// 1~570 -> /initializeで注入されるseason1期間の配信
		go simulateRandomLivestreamViewer(ctx, client, loginUsers[userIdx+1], 1 /* livestream id start */, 570 /* livestream id end*/, "Season1")
	}

	<-ctx.Done()
	log.Println("season1 user workers has finished.")

	return nil
}
