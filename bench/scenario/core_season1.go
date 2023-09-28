package scenario

import (
	"context"
	"errors"
	"log"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/generator"
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

// userID -> user
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
func Season1(ctx context.Context, webappIPAddress string) {
	log.Println("running season1 scenario ...")

	// 広告費用で制御して、リクエスト送信goroutineを単純倍増
	// INFO: リクエスト数を制御するだけでなく、tipsの金額も増加させても良いかもしれない
	for userIdx := 0; userIdx < config.AdvertiseCost; userIdx++ {
		// 1~570 -> /initializeで注入されるseason1期間の配信
		go simulateRandomLivestreamViewer(ctx, webappIPAddress, loginUsers[userIdx+1], 1 /* livestream id start */, 570 /* livestream id end*/, "Season1")
	}

	<-ctx.Done()
	log.Println("season1 user workers has finished.")
}

func simulateSeason1User(ctx context.Context, webappIPAddress string, loginUser loginUser) {
	client, err := isupipe.NewClient(
		agent.WithBaseURL(webappIPAddress),
	)
	if err != nil {
		panic(err)
	}

	// FIXME: 自然なリクエストにするためには、複数のユーザからリクエストが飛んでほしい
	//        isupipe.Clientのログインセッションキャッシュを考慮しつつ、
	//        season1 scenario内で複数のgoroutineを吐き出して、それぞれのユーザをシミュレートするように変更する
	loginRequest := isupipe.LoginRequest{
		UserName: loginUser.UserName,
		Password: loginUser.Password,
	}

	if err := client.Login(ctx, &loginRequest); err != nil {
		log.Printf("reaction: failed to login: %s\n", err.Error())
		return
	}

	season1UserWorker, err := worker.NewWorker(func(ctx context.Context, i int) {
		postReactionReq := isupipe.PostReactionRequest{
			EmojiName: generator.GenerateRandomReaction(),
		}

		randomLivestreamID := generator.GenerateIntBetween(1, 11)
		postedReaction, err := client.PostReaction(ctx, randomLivestreamID /* livestream id*/, &postReactionReq)
		if err != nil {
			if errors.Is(err, isupipe.ErrCancelRequest) {
				return
			} else {
				log.Printf("reaction: failed to post reaction : %s\n", err.Error())
				return
			}
		}

		// ちゃんと結果整合性が担保されているかチェック
		if err := assertPostedReactionConsistency(ctx, client, randomLivestreamID, postedReaction.ID); err != nil {
			bencherror.WrapError(bencherror.DBInconsistencyError, err)
			// log.Printf("Season: %s\n", err)
		}

		// season1でたまたま高額Tipが連続すると、すぐに条件を達成してしまう
		// ある程度のリクエストをさばけることを検証するべく、tip-levelをおさえこむ
		// TipLevel1であれば、最高でも500で、200kまでに4000リクエストを要するため、一旦そうしておく
		// randomTipLevel := generator.GenerateRandomTipLevel()
		postSuperchatReq := isupipe.PostSuperchatRequest{
			Comment: generator.GenerateRandomComment(),
			Tip:     generator.GenerateTip(generator.TipLevel1),
		}
		_, err = client.PostSuperchat(ctx, randomLivestreamID /* livestream id*/, &postSuperchatReq)
		if err != nil {
			if errors.Is(err, isupipe.ErrCancelRequest) {
				return
			} else {
				log.Printf("reaction: failed to post superchat : %s\n", err.Error())
				return
			}
		}

		// ちゃんと結果整合性が担保されているかチェック
		// if err := assertPostedSuperchatConsistency(ctx, client, randomLivestreamID, postedSuperchat.Id); err != nil {
		// 	bencherror.WrapError(bencherror.DBInconsistencyError, err)
		// 	log.Printf("Season: %s\n", err)
		// }
	}, worker.WithInfinityLoop())
	if err != nil {
		log.Printf("WARNING: found an error; Season1 scenario does not anything: %s\n", err.Error())
		return
	}
	season1UserWorker.SetParallelism(config.DefaultBenchmarkerParallelism)

	log.Println("processing workers ...")
	season1UserWorker.Process(ctx)
	<-ctx.Done()
	season1UserWorker.Wait()

}
