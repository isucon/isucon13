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

// simulateRandomLivestreamViewer は特定の配信に対してスーパーチャットとリアクションを送信し続けるWorkerを動かす
// randomViewLivestreamIDStart~randomViewLivestreamIDEnd の範囲内で特定の配信を選出し、その配信に対してスパチャ/リアクションする
func simulateRandomLivestreamViewer(
	ctx context.Context,
	webappIPAddress string,
	loginUser loginUser,
	randomViewLivestreamIDStart int,
	randomViewLivestreamIDEnd int,
	scenarioName string,
) {
	client, err := isupipe.NewClient(
		agent.WithBaseURL(webappIPAddress),
	)
	if err != nil {
		panic(err)
	}

	loginRequest := isupipe.LoginRequest{
		UserName: loginUser.UserName,
		Password: loginUser.Password,
	}

	if err := client.Login(ctx, &loginRequest); err != nil {
		log.Printf("%s: %s\n", scenarioName, err.Error())
		return
	}

	userSimulateWorker, err := worker.NewWorker(func(ctx context.Context, i int) {

		randomLivestreamID := generator.GenerateIntBetween(randomViewLivestreamIDStart, randomViewLivestreamIDEnd)

		if err := client.EnterLivestream(ctx, randomLivestreamID /* livestream id*/); err != nil {
			log.Printf("%s: %s\n", scenarioName, err.Error())
			return
		}

		postReactionReq := isupipe.PostReactionRequest{
			EmojiName: generator.GenerateRandomReaction(),
		}
		postedReaction, err := client.PostReaction(ctx, randomLivestreamID /* livestream id*/, &postReactionReq)
		if err != nil {
			log.Printf("%s: %s\n", scenarioName, err.Error())
			return
		}

		// ちゃんと結果整合性が担保されているかチェック
		if err := checkPostedReactionConsistency(ctx, client, randomLivestreamID, postedReaction.ID); err != nil {
			log.Printf("%s: %s\n", scenarioName, err.Error())
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
			log.Printf("%s: %s\n", scenarioName, err.Error())
			return
		}

		// ちゃんと結果整合性が担保されているかチェック
		if err := checkPostedSuperchatConsistency(ctx, client, randomLivestreamID, postedSuperchat.ID); err != nil {
			log.Printf("%s: %s\n", scenarioName, err.Error())
		}

		if err := client.LeaveLivestream(ctx, randomLivestreamID /* livestream id*/); err != nil {
			log.Printf("%s: %s\n", scenarioName, err.Error())
			return
		}

		// INFO: ここで適当なsleepを入れて、広告費用によってsleep間隔が狭まるようにしてもいいかも
		//       goroutineの生成数はマシンの影響を強く受けるのと、リクエストの多様性が損なわれる
	}, worker.WithInfinityLoop())
	if err != nil {
		log.Printf("WARNING: found an error; %s scenario does not anything: %s\n", scenarioName, err.Error())
		return
	}
	userSimulateWorker.SetParallelism(config.DefaultBenchmarkerParallelism)

	log.Println("processing workers ...")
	userSimulateWorker.Process(ctx)
	<-ctx.Done()
	userSimulateWorker.Wait()

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

	postedReactionFound := false
	for _, r := range reactions {
		if r.ID == postedReactionID {
			postedReactionFound = true
			break
		}
	}

	if !postedReactionFound {
		err := fmt.Errorf("投稿されたリアクション(id: %d)が取得できませんでした", postedReactionID)
		return bencherror.DBInconsistency(err)
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

	postedSuperchatFound := false
	for _, s := range superchats {
		if s.ID == postedSuperchatID {
			postedSuperchatFound = true
			break
		}
	}

	if !postedSuperchatFound {
		err := fmt.Errorf("投稿されたスーパーチャット(id: %d)が取得できませんでした", postedSuperchatID)
		return bencherror.DBInconsistency(err)
	}

	return nil
}
