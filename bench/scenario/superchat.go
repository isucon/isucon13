package scenario

import (
	"context"
	"log"

	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/generator"
	"github.com/isucon/isucon13/bench/isupipe"
)

// Superchat は基本機能のベンチマークを実行する
// INFO: 後々シーズンごとのシナリオに移行される、一時的なシナリオ
func Superchat(ctx context.Context, client *isupipe.Client) {
	log.Println("running superchat scenario ...")

	// init.sqlで事前挿入されたデータ
	loginRequest := isupipe.LoginRequest{
		UserName: "井上 太郎",
		Password: "o^E0K1Axj@",
	}
	if err := client.Login(ctx, &loginRequest); err != nil {
		// log.Printf("Superchat: failed to login: %s\n", err.Error())
		return
	}

	postSuperchatWorker, err := worker.NewWorker(func(ctx context.Context, i int) {

		// log.Printf("worker %d posting superchat request ...\n", i)
		req := isupipe.PostSuperchatRequest{
			Comment: generator.GenerateRandomComment(),
			Tip:     0, // superchatシナリオでは常にtips == 0
		}

		if _, err := client.PostSuperchat(ctx, 1 /* livestream id*/, &req); err != nil {
			// log.Printf("Superchat: failed to post superchat: %s\n", err.Error())
			return
		}
	}, worker.WithInfinityLoop())
	if err != nil {
		log.Printf("WARNING: found an error; Superchat scenario does not anything: %s\n", err.Error())
		return
	}
	postSuperchatWorker.SetParallelism(config.DefaultBenchmarkerParallelism)

	log.Println("processing workers ...")
	postSuperchatWorker.Process(ctx)

	log.Println("waiting context canceling ...")
	<-ctx.Done()
	log.Println("waiting for post superchat workers ...")
	postSuperchatWorker.Wait()
	log.Println("post superchat workers has finished.")
}
