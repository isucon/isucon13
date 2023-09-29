package scenario

import (
	"context"
	"errors"
	"log"

	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/generator"
	"github.com/isucon/isucon13/bench/isupipe"
)

// Tips は投げ銭機能のベンチマークを実行する
// INFO: 後々シーズンごとのシナリオに移行される、一時的なシナリオ
func Tips(ctx context.Context, client *isupipe.Client) {
	log.Println("running tips scenario ...")

	// 事前挿入されたデータ
	loginRequest := isupipe.LoginRequest{
		UserName: "高橋 智也",
		Password: "Ba)6J7pmZY",
	}
	if err := client.Login(ctx, &loginRequest); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return
		}
		log.Printf("Tips: %s\n", err.Error())
		return
	}

	if err := client.EnterLivestream(ctx, 1 /* livestream id*/); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return
		}
		log.Printf("Tips: %s\n", err.Error())
	}

	postLivecommentWorker, err := worker.NewWorker(func(ctx context.Context, i int) {
		// log.Printf("worker %d posting tips request ...\n", i)
		randomTipLevel := generator.GenerateRandomTipLevel()
		req := isupipe.PostLivecommentRequest{
			Comment: generator.GenerateRandomComment(),
			Tip:     generator.GenerateTip(randomTipLevel), // livecommentシナリオでは常にtips == 0
		}

		if _, err := client.PostLivecomment(ctx, 1 /* livestream id*/, &req); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return
			}
			log.Printf("Tips: %s\n", err.Error())
			return
		}
	}, worker.WithInfinityLoop())
	if err != nil {
		log.Printf("WARNING: found an error; Tips scenario does not anything: %s\n", err.Error())
		return
	}
	postLivecommentWorker.SetParallelism(config.DefaultBenchmarkerParallelism)

	log.Println("processing workers ...")
	postLivecommentWorker.Process(ctx)

	log.Println("waiting context canceling ...")
	<-ctx.Done()
	log.Println("waiting for post livecomment workers ...")
	postLivecommentWorker.Wait()
	log.Println("post livecomment workers has finished.")

	if err := client.LeaveLivestream(ctx, 1 /* livestream id*/); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return
		}
		log.Printf("Tips: %s\n", err.Error())
	}
}
