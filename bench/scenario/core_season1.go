package scenario

import (
	"context"
	"log"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/generator"
	"github.com/isucon/isucon13/bench/isupipe"
)

type loginUser struct {
	UserName string
	Password string
}

var loginUsers = []loginUser{
	{
		UserName: "鈴木 陽一",
		Password: "9!5AgcWmQ@",
	},
	{
		UserName: "山本 拓真",
		Password: "gIfYB5Oqm+",
	},
	{
		UserName: "山下 陽子",
		Password: ")3AaHkiCa7",
	},
	{
		UserName: "藤井 京助",
		Password: "ujO08RwS_6",
	},
	{
		UserName: "井上 淳",
		Password: "8_HZfg1s$C",
	},
	{
		UserName: "遠藤 淳",
		Password: "8k4Xx)9wg^",
	},
	{
		UserName: "田中 学",
		Password: "33DEv!_#_p",
	},
	{
		UserName: "三浦 翔太",
		Password: "0*opW%Kp!j",
	},
	{
		UserName: "佐藤 くみ子",
		Password: "%86BURTp69",
	},
	{
		UserName: "池田 京助",
		Password: "cJ&2Ow*gnk",
	},
}

// Season1 シナリオは、サービス開始時点で存在する配信者の配信に対して、ランダムにリクエストを送信する
func Season1(ctx context.Context, webappIPAddress string) {
	log.Println("running season1 scenario ...")

	for _, user := range loginUsers {
		go simulateSeason1User(ctx, webappIPAddress, user)
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
		// log.Printf("reaction: failed to login: %s\n", err.Error())
		return
	}

	season1UserWorker, err := worker.NewWorker(func(ctx context.Context, i int) {
		postReactionReq := isupipe.PostReactionRequest{
			EmojiName: generator.GenerateRandomReaction(),
		}

		randomLivestreamID := generator.GenerateIntBetween(1, 11)
		if err := client.PostReaction(ctx, randomLivestreamID /* livestream id*/, &postReactionReq); err != nil {
			// log.Printf("reaction: failed to post reaction : %s\n", err.Error())
			return
		}
		benchscore.AddScore(benchscore.SuccessPostReaction)

		randomTipLevel := generator.GenerateRandomTipLevel()
		postSuperchatReq := isupipe.PostSuperchatRequest{
			Comment: generator.GenerateRandomComment(),
			Tip:     generator.GenerateTip(randomTipLevel),
		}
		if _, err := client.PostSuperchat(ctx, randomLivestreamID /* livestream id*/, &postSuperchatReq); err != nil {
			// log.Printf("reaction: failed to post reaction : %s\n", err.Error())
			return
		}
		benchscore.AddScore(benchscore.SuccessPostSuperchat)
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
