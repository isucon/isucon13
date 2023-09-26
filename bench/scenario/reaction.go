package scenario

import (
	"context"
	"log"

	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/generator"
	"github.com/isucon/isucon13/bench/isupipe"
)

// INFO: 後々シーズンごとのシナリオに移行される、一時的なシナリオ
func Reaction(ctx context.Context, client *isupipe.Client) {
	log.Println("running reaction scenario ...")
	postReactionWorker, err := worker.NewWorker(func(ctx context.Context, i int) {
		// init.sqlで事前挿入されたデータ
		loginRequest := isupipe.LoginRequest{
			UserName: "鈴木 陽一",
			Password: "kaorisuzuki",
		}
		if err := client.Login(ctx, &loginRequest); err != nil {
			// log.Printf("reaction: failed to login: %s\n", err.Error())
			return
		}

		// log.Printf("worker %d posting reaction request ...\n", i)
		req := isupipe.PostReactionRequest{
			EmojiName: generator.GenerateRandomReaction(),
		}

		if err := client.PostReaction(ctx, 1 /* livestream id*/, &req); err != nil {
			// log.Printf("reaction: failed to post reaction : %s\n", err.Error())
			return
		}

		benchscore.AddScore(benchscore.SuccessPostReaction)
	}, worker.WithInfinityLoop())
	if err != nil {
		log.Printf("WARNING: found an error; Reaction scenario does not anything: %s\n", err.Error())
		return
	}
	postReactionWorker.SetParallelism(config.DefaultBenchmarkerParallelism)

	log.Println("processing workers ...")
	postReactionWorker.Process(ctx)

	log.Println("waiting context canceling ...")
	<-ctx.Done()
	log.Println("waiting for post reaction workers ...")
	postReactionWorker.Wait()
	log.Println("post reaction workers has finished.")
}
