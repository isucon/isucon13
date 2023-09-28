package scenario

import (
	"context"
	"log"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
)

// Season2 シナリオは、新人配信者の予約が発生しつつ、Season1と同様に配信に対するスパチャ/投げ銭を行う
func Season2(ctx context.Context, webappIPAddress string) {
	log.Println("running season2 scenario ...")

	// ログイン回数を減らしてベンチマーカの性能を上げるため、
	// ログイン済みのクライアントをキャッシュする

	userIDToClient := map[int]*isupipe.Client{}

	for userID, user := range loginUsers {
		client, err := isupipe.NewClient(
			agent.WithBaseURL(webappIPAddress),
		)
		if err != nil {
			bencherror.Internal(err)
			return
		}

		loginRequest := isupipe.LoginRequest{
			UserName: user.UserName,
			Password: user.Password,
		}
		if err := client.Login(ctx, &loginRequest); err != nil {
			log.Printf("season2: %s\n", err.Error())
			return
		}

		userIDToClient[userID] = client
	}

	season2ReserveWorker, err := worker.NewWorker(func(ctx context.Context, i int) {
		reservePattern := scheduler.Season2LivestreamReservationPatterns[i]
		client := userIDToClient[reservePattern.UserID]

		reserveRequest := isupipe.ReserveLivestreamRequest{
			Title:         reservePattern.Title,
			Description:   reservePattern.Description,
			PrivacyStatus: "public",
			StartAt:       reservePattern.StartAt.Unix(),
			EndAt:         reservePattern.EndAt.Unix(),
		}
		if _, err := client.ReserveLivestream(ctx, &reserveRequest); err != nil {
			log.Printf("season2: %s\n", err.Error())
		}
	}, worker.WithLoopCount(int32(len(scheduler.Season2LivestreamReservationPatterns))))
	if err != nil {
		log.Printf("WARNING: found an error; Season1 scenario does not anything: %s\n", err.Error())
		return
	}
	season2ReserveWorker.SetParallelism(config.DefaultBenchmarkerParallelism)

	log.Println("processing workers ...")
	season2ReserveWorker.Process(ctx)
	// 予約リクエストを捌ききれないと、スパチャ/リアクションが投げられないようにする
	season2ReserveWorker.Wait()

	// season1と同様に広告費用で制御して、リクエスト送信goroutineを単純倍増
	// INFO: リクエスト数を制御するだけでなく、tipsの金額も増加させても良いかもしれない
	for userIdx := 0; userIdx < config.AdvertiseCost; userIdx++ {
		// 571~(571+len(Season2LivestreamReservationPatterns)) -> season1期間の配信を含まない、season2ReserveWorkerが登録する配信一覧
		randomLivestreamIDStartAt := Season1GeneratedLivestreamCount + 1
		randomLivestreamIDEndAt := Season1GeneratedLivestreamCount + 1 + len(scheduler.Season2LivestreamReservationPatterns)
		go simulateRandomLivestreamViewer(ctx, webappIPAddress, loginUsers[userIdx+1], randomLivestreamIDStartAt, randomLivestreamIDEndAt, "Season2")
	}

	<-ctx.Done()
	log.Println("season2 workers has finished.")
}
