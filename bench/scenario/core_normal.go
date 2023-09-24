package scenario

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/isupipe"
)

// Normal は基本機能のベンチマークを実行する
// INFO: 後々シーズンごとのシナリオに移行される、一時的なシナリオ
func Normal(ctx context.Context, client *isupipe.Client) {
	log.Println("running normal scenario ...")
	postSuperchatWorker, err := worker.NewWorker(func(ctx context.Context, i int) {
		// init.sqlで事前挿入されたデータ
		loginRequest := isupipe.LoginRequest{
			UserName: "isupipe",
			Password: "1sup1pe",
		}
		if err := client.Login(ctx, &loginRequest); err != nil {
			log.Printf("Normal: failed to login: %s\n", err.Error())
			return
		}

		log.Printf("worker %d posting superchat request ...\n", i)
		req := isupipe.PostSuperchatRequest{
			Comment: fmt.Sprintf("%d", i),
		}

		if _, err := client.PostSuperchat(ctx, 1 /* livestream id*/, &req); err != nil {
			log.Printf("Normal: failed to post superchat: %s\n", err.Error())
			return
		}
	}, worker.WithInfinityLoop())
	if err != nil {
		log.Printf("WARNING: found an error; Normal scenario does not anything: %s\n", err.Error())
		return
	}
	postSuperchatWorker.SetParallelism(config.DefaultBenchmarkerParallelism)

	log.Println("processing workers ...")
	workerCtx, cancelWorkerCtx := context.WithTimeout(ctx, config.DefaultBenchmarkWorkerTimeout*time.Second)
	defer cancelWorkerCtx()
	postSuperchatWorker.Process(workerCtx)

	log.Println("waiting context canceling ...")
	<-workerCtx.Done()
	log.Println("waiting for post superchat workers ...")
	postSuperchatWorker.Wait()
	log.Println("post superchat workers has finished.")
}
