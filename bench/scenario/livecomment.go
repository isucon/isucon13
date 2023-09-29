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

// Livecomment は基本機能のベンチマークを実行する
// INFO: 後々シーズンごとのシナリオに移行される、一時的なシナリオ
func Livecomment(ctx context.Context, client *isupipe.Client) {
	log.Println("running livecomment scenario ...")

	// init.sqlで事前挿入されたデータ
	loginRequest := isupipe.LoginRequest{
		UserName: "井上 太郎",
		Password: "o^E0K1Axj@",
	}
	if err := client.Login(ctx, &loginRequest); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return
		}
		log.Printf("Livecomment: %s\n", err.Error())
		// log.Printf("Livecomment: failed to login: %s\n", err.Error())
		return
	}

	if err := client.EnterLivestream(ctx, 1 /* livestream id*/); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return
		}
		log.Printf("Livecomment: %s\n", err.Error())
	}

	postLivecommentWorker, err := worker.NewWorker(func(ctx context.Context, i int) {

		// log.Printf("worker %d posting livecomment request ...\n", i)
		req := isupipe.PostLivecommentRequest{
			Comment: generator.GenerateRandomComment(),
			Tip:     0, // livecommentシナリオでは常にtips == 0
		}

		if _, err := client.PostLivecomment(ctx, 1 /* livestream id*/, &req); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return
			}
			log.Printf("Livecomment: %s\n", err.Error())
			return
		}

	}, worker.WithInfinityLoop())
	if err != nil {
		log.Printf("WARNING: found an error; Livecomment scenario does not anything: %s\n", err.Error())
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
		log.Printf("Livecomment: %s\n", err.Error())
	}
}
