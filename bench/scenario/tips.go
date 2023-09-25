package scenario

import (
	"context"
	"log"
	"time"

	"github.com/isucon/isucandar/worker"
	"github.com/isucon/isucon13/bench/internal/benchscore"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/generator"
	"github.com/isucon/isucon13/bench/isupipe"
)

// Tips は投げ銭機能のベンチマークを実行する
// INFO: 後々シーズンごとのシナリオに移行される、一時的なシナリオ
func Tips(ctx context.Context, client *isupipe.Client) {
	log.Println("running tips scenario ...")
	postSuperchatWorker, err := worker.NewWorker(func(ctx context.Context, i int) {
		// init.sqlで事前挿入されたデータ
		loginRequest := isupipe.LoginRequest{
			UserName: "isupipe",
			Password: "1sup1pe",
		}
		if err := client.Login(ctx, &loginRequest); err != nil {
			log.Printf("Superchat: failed to login: %s\n", err.Error())
			return
		}

		// log.Printf("worker %d posting tips request ...\n", i)
		randomTipLevel := generator.GenerateRandomTipLevel()
		req := isupipe.PostSuperchatRequest{
			Comment: generator.GenerateRandomComment(),
			Tip:     generator.GenerateTip(randomTipLevel), // superchatシナリオでは常にtips == 0
		}

		if _, err := client.PostSuperchat(ctx, 1 /* livestream id*/, &req); err != nil {
			// log.Printf("Tips: failed to post superchat: %s\n", err.Error())
			return
		}

		benchscore.AddTipProfit(req.Tip)
	}, worker.WithInfinityLoop())
	if err != nil {
		log.Printf("WARNING: found an error; Tips scenario does not anything: %s\n", err.Error())
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
