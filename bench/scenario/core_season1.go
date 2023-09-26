package scenario

import (
	"context"
	"fmt"
	"log"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon13/bench/internal/bencherror"
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
		UserName: "渡辺 裕美子",
		Password: ")fi8AoKoed",
	},
	{
		UserName: "佐藤 翔太",
		Password: "!4+3O*Tv&$",
	},
	{
		UserName: "渡辺 くみ子",
		Password: "kCG9Vio#@0",
	},
	{
		UserName: "清水 陽一",
		Password: "R2QbFf@n)G",
	},
	{
		UserName: "斉藤 裕美子",
		Password: "vcBK18OI%&",
	},
	{
		UserName: "鈴木 舞",
		Password: "*R++jCXz+3",
	},
	{
		UserName: "井上 くみ子",
		Password: "K4RJaB7y#g",
	},
	{
		UserName: "松田 淳",
		Password: "z(U58NunYm",
	},
	{
		UserName: "太田 京助",
		Password: "^h9R^E^h2I",
	},
	{
		UserName: "伊藤 修平",
		Password: "@X0$0DMoCc",
	},
}

// Season1 シナリオは、サービス開始時点で存在する配信者の配信に対して、ランダムにリクエストを送信する
func Season1(ctx context.Context, webappIPAddress string) {
	log.Println("running season1 scenario!!!! ...")

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

	season1UserWorker, err := worker.NewWorker(func(ctx context.Context, i int) {
		// FIXME: 自然なリクエストにするためには、複数のユーザからリクエストが飛んでほしい
		//        isupipe.Clientのログインセッションキャッシュを考慮しつつ、
		//        season1 scenario内で複数のgoroutineを吐き出して、それぞれのユーザをシミュレートするように変更する
		loginRequest := isupipe.LoginRequest{
			UserName: loginUser.UserName,
			Password: loginUser.Password,
		}

		log.Printf("login: username='%s', password='%s'\n", loginUser.UserName, loginUser.Password)
		if err := client.Login(ctx, &loginRequest); err != nil {
			// log.Printf("reaction: failed to login: %s\n", err.Error())
			return
		}

		postReactionReq := isupipe.PostReactionRequest{
			EmojiName: generator.GenerateRandomReaction(),
		}

		randomLivestreamID := generator.GenerateIntBetween(1, 11)
		postedReaction, err := client.PostReaction(ctx, randomLivestreamID /* livestream id*/, &postReactionReq)
		if err != nil {
			// log.Printf("reaction: failed to post reaction : %s\n", err.Error())
			return
		}

		if err := checkPostedReactionConsistency(ctx, client, randomLivestreamID, postedReaction.ID); err != nil {
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
		postedSuperchat, err := client.PostSuperchat(ctx, randomLivestreamID /* livestream id*/, &postSuperchatReq)
		if err != nil {
			// log.Printf("reaction: failed!!! to post reaction : %s\n", err.Error())
			return
		}

		// ちゃんと結果整合性が担保されているかチェック
		if err := checkPostedSuperchatConsistency(ctx, client, randomLivestreamID, postedSuperchat.Id); err != nil {
			bencherror.WrapError(bencherror.DBInconsistencyError, err)
			// log.Printf("Season: %s\n", err)
		}
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

func checkPostedReactionConsistency(
	ctx context.Context,
	client *isupipe.Client,
	livestreamID int,
	postedReactionID int,
) error {
	reactions, err := client.GetReactions(ctx, livestreamID)
	if err != nil {
		return err
	}

	var postedReaction *isupipe.Reaction
	for _, r := range reactions {
		if r.ID == postedReactionID {
			*postedReaction = r
		}
	}

	if postedReaction == nil {
		return fmt.Errorf("投稿されたリアクション(id: %d)が取得できませんでした", postedReactionID)
	}

	return nil
}

func checkPostedSuperchatConsistency(
	ctx context.Context,
	client *isupipe.Client,
	livestreamID int,
	postedSuperchatID int,
) error {
	superchats, err := client.GetSuperchats(ctx, livestreamID)
	if err != nil {
		return err
	}

	var postedSuperchat *isupipe.Superchat
	for _, s := range superchats {
		if s.ID == postedSuperchatID {
			*postedSuperchat = s
		}
	}

	if postedSuperchat == nil {
		return fmt.Errorf("投稿されたスーパーチャット(id: %d)が取得できませんでした", postedSuperchatID)
	}

	return nil
}
