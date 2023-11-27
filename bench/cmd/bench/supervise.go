package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

// FIXME: SQSのメッセージサイズが最大で256KBなので、200KB程度までで打ち切るように

var (
	accessKey, secretAccessKey string
	slackWebhookURL            string
	messageLimit               int
	production                 bool
)

const (
	MsgTimeout = "ベンチマーク処理がタイムアウトしました"
	MsgFail    = "運営に連絡してください"
)

const (
	StatusRunning = "running"
	StatusSuccess = "done"
	StatusFailed  = "aborted"
	StatusTimeout = "aborted"
)

var AZName string

func init() {

}

func ResolveAZName(azID string) (string, bool) {
	m := map[string]string{
		"ap-northeast-1a": "apne1-az4",
		"ap-northeast-1c": "apne1-az1",
		"ap-northeast-1d": "apne1-az2",
	}

	azName, ok := m[azID]
	return azName, ok
}

type TaskMetadataV4 struct {
	AvailabilityZone string `json:"AvailabilityZone"`
}

func fetchAZName(ctx context.Context) (string, error) {
	metadataBaseUrl := os.Getenv("ECS_CONTAINER_METADATA_URI_V4")
	if metadataBaseUrl == "" {
		return "", fmt.Errorf("empty metadata url")
	}

	metadataUrl, err := url.JoinPath(metadataBaseUrl, "task")
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, metadataUrl, nil)
	if err != nil {
		return "", err
	}
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	taskMetadata := &TaskMetadataV4{}
	if err := json.NewDecoder(resp.Body).Decode(&taskMetadata); err != nil {
		return "", err
	}

	azName, ok := ResolveAZName(taskMetadata.AvailabilityZone)
	if !ok {
		return "", fmt.Errorf("failed to resolve availability zone name for %s", taskMetadata.AvailabilityZone)
	}

	return azName, nil
}

// messagesの末尾からn行を結合して取得
func joinN(messages []string, n int) string {
	if len(messages) <= n {
		return strings.Join(messages, "\n")
	}
	return strings.Join(messages[len(messages)-n:], "\n")
}

func execBench(ctx context.Context, job *Job) (*Result, error) {
	target := job.TargetIP
	executablePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	log.Printf("executable = %s\n", executablePath)

	// ベンチマーカー実行前に確実にresultファイルを削除する
	// 他のチームに対する結果が含まれている可能性があるため
	log.Println("cleanup old logs for current job")
	for _, name := range []string{config.ResultPath, config.StaffLogPath, config.ContestantLogPath} {
		os.Remove(name)
	}

	benchOptions := []string{
		"run",
		"--nameserver", target,
		"--staff-log-path", config.StaffLogPath,
		"--contestant-log-path", config.ContestantLogPath,
		"--result-path", config.ResultPath,
	}
	if production {
		benchOptions = append(benchOptions, "--enable-ssl")
		benchOptions = append(benchOptions, "--target", "https://pipe.u.isucon.dev:443")
	} else {
		// NOTE: 開発環境(Docker)
		// benchOptions = append(benchOptions, "--dns-port", strconv.Itoa(1053))
		benchOptions = append(benchOptions, "--enable-ssl")
		benchOptions = append(benchOptions, "--target", "https://pipe.u.isucon.dev:443")
	}
	for _, server := range job.Servers {
		benchOptions = append(benchOptions, "--webapp")
		benchOptions = append(benchOptions, server)
	}
	log.Printf("benchmark options = %+v\n", benchOptions)

	var stdout, stderr bytes.Buffer
	// 余裕みて3分
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, executablePath, benchOptions...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = 13 * time.Second

	status := StatusSuccess

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		log.Println("running benchmark ...")
		if err := cmd.Run(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("execBench中断")
		NotifyWorkerErr(job, ctx.Err(), stdout.String(), stderr.String(), "ベンチマーカーの実行がタイムアウトしました (StatusFailed)")
		status = StatusTimeout
	case err, ok := <-errCh:
		if ok && err != nil {
			log.Printf("execBenchでエラー発生: %s\n", err.Error())
			NotifyWorkerErr(job, err, stdout.String(), stderr.String(), "ベンチマーカーの実行エラーが発生 (StatusFailed)")
			status = StatusFailed
		}
	}

	var msgs []string
	log.Println("read contestant log path")
	// stdout
	b, err := os.ReadFile(config.ContestantLogPath)
	if err != nil {
		return &Result{
			ID:         job.ID,
			Stdout:     joinN(strings.Split(stdout.String(), "\n"), messageLimit),
			Stderr:     joinN(strings.Split(stderr.String(), "\n"), messageLimit),
			Reason:     err.Error(),
			IsPassed:   false,
			Score:      0,
			Status:     status,
			FinishedAt: time.Now(),
		}, nil
	}
	contestantLog := strings.Split(string(b), "\n")

	log.Println("read result path")
	// result
	var benchResult *BenchResult
	b, err = os.ReadFile(config.ResultPath)
	if err != nil {
		return &Result{
			ID:         job.ID,
			Stdout:     joinN(strings.Split(stdout.String(), "\n"), messageLimit),
			Stderr:     joinN(strings.Split(stderr.String(), "\n"), messageLimit),
			Reason:     err.Error(),
			IsPassed:   false,
			Score:      0,
			Status:     status,
			FinishedAt: time.Now(),
		}, nil
	}

	log.Println("decode bench result")
	//
	if err := json.NewDecoder(bytes.NewBuffer(b)).Decode(&benchResult); err != nil {
		return &Result{
			ID:         job.ID,
			Stdout:     joinN(strings.Split(stdout.String(), "\n"), messageLimit),
			Stderr:     joinN(strings.Split(stderr.String(), "\n"), messageLimit),
			Reason:     err.Error(),
			IsPassed:   false,
			Score:      0,
			Status:     status,
			FinishedAt: time.Now(),
		}, nil
	}

	msgs = contestantLog
	msgs = append(msgs, benchResult.Messages...)

	if status == StatusSuccess {
		log.Println("success benchmark")
		return &Result{
			ID:            job.ID,
			Stdout:        joinN(strings.Split(stdout.String(), "\n"), messageLimit),
			Stderr:        joinN(strings.Split(stderr.String(), "\n"), messageLimit),
			Reason:        joinN(msgs, messageLimit),
			IsPassed:      benchResult.Pass,
			Score:         benchResult.Score,
			ResolvedCount: benchResult.ResolvedCount,
			Language:      benchResult.Language,
			Status:        status,
			FinishedAt:    time.Now(),
		}, nil
	} else {
		log.Println("fail benchmark")
		return &Result{
			ID:         job.ID,
			Stdout:     joinN(strings.Split(stdout.String(), "\n"), messageLimit),
			Stderr:     joinN(strings.Split(stderr.String(), "\n"), messageLimit),
			Reason:     "ベンチマーク失敗",
			IsPassed:   false,
			Score:      0,
			Status:     status,
			FinishedAt: time.Now(),
		}, nil
	}
}

var supervise = cli.Command{
	Name:  "supervise",
	Usage: "supervisor実行",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "access-key",
			Value:       "",
			Destination: &accessKey,
			EnvVar:      "SUPERVISOR_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:        "secret-access-key",
			Value:       "",
			Destination: &secretAccessKey,
			EnvVar:      "SUPERVISOR_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:        "slack-webhook-url",
			Value:       "",
			Destination: &slackWebhookURL,
			EnvVar:      "SUPERVISOR_SLACK_WEBHOOK_URL",
		},
		cli.IntFlag{
			Name:        "message-limit",
			Value:       200,
			Destination: &messageLimit,
			EnvVar:      "SUPERVISOR_MESSAGE_LIMIT",
		},
		cli.BoolFlag{
			Name:        "production",
			Destination: &production,
			EnvVar:      "SUPERVISOR_PRODUCTION",
		},
	},
	Action: func(cliCtx *cli.Context) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
		defer cancel()
		log.Println("Start ISUPipe Supervisor")

		log.Println("Fetching AZ Name ...")
		azName, err := fetchAZName(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
		AZName = azName
		log.Printf("AZ Name = %s\n", AZName)

		privateKey, err := os.ReadFile("/home/benchuser/cmd/bench/id_ed25519")
		if err != nil {
			log.Printf("privateKey error = %s\n", err)
			return err
		}
		signer, err := ssh.ParsePrivateKey(privateKey)
		if err != nil {
			log.Printf("signer error = %s\n", err)
			return err
		}

		var portal *Portal
		var finalcheckBucketName string
		if production {
			log.Println("Running on production")
			// NOTE: アクセスキーを上書き
			accessKey = ""
			secretAccessKey = ""
			// NOTE: SQSはsqs初期化時に決まる
			finalcheckBucketName = "isucon13-finalcheck-prod"

			portal, err = NewPortal(
				AZName,
				Production,
				accessKey,
				secretAccessKey,
			)
			if err != nil {
				log.Println("failed to initiate portal")
				return cli.NewExitError(err, 1)
			}
		} else {
			log.Println("Running on development")
			finalcheckBucketName = "isucon13-finalcheck-dev"

			portal, err = NewPortal(
				AZName,
				Develop,
				accessKey,
				secretAccessKey,
			)
			if err != nil {
				log.Println("failed to initiate portal")
				return cli.NewExitError(err, 1)
			}
		}
		jobCh := portal.StartReceiveJob(ctx)

		for {
			select {
			case <-ctx.Done():
				return cli.NewExitError(ctx.Err(), 1)
			case job := <-jobCh:
				log.Printf("receive job = %+v\n", job)

				if job.Action == "reboot" {
					log.Println("Job is reboot task")
					var errs []string
					for _, server := range job.Servers {
						if err := reboot(server, signer); err != nil {
							errs = append(errs, err.Error())
						}
					}
					if len(errs) > 0 {
						NotifyWorkerErr(job, nil, "", "", strings.Join(errs, ","))
					}
					log.Println("reboot completed")
				} else {
					log.Println("Job is benchmark")

					log.Println("change status running")
					if err := portal.SendResult(ctx, job, NewRunningResult(job.ID)); err != nil {
						NotifyWorkerErr(job, err, "", "", "ベンチマーカーの実行に失敗。すぐに調査してください。supervisorの処理は継続します")
					}

					log.Println("execute benchmark")
					result, err := execBench(ctx, job)
					if err != nil {
						NotifyWorkerErr(job, err, "", "", "ベンチマーカーの実行に失敗。すぐに調査してください。supervisorの処理は継続します")
					}

					log.Println("report result")
					if err := portal.SendResult(ctx, job, result); err != nil {
						NotifyWorkerErr(job, err, "", "", "ベンチマーカーの結果送信に失敗。すぐに調査してください。supervisorの処理は継続します")
					}

					_ = finalcheckBucketName
					// log.Println("upload finalcheck result")
					// if err := UploadFinalcheckResult(finalcheckBucketName, job.ID, job.Team); err != nil {
					// 	NotifyWorkerErr(job, err, "", "", "FinalCheckの結果送信に失敗。すぐに調査してください。supervisorの処理は継続します")
					// }

					log.Println("cleanup old logs for next job")
					os.Remove(config.StaffLogPath)
					os.Remove(config.ContestantLogPath)
					os.Remove(config.ResultPath)
				}
			}
		}
	},
}
