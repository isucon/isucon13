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

	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
)

// FIXME: SQSのメッセージサイズが最大で256KBなので、200KB程度までで打ち切るように

const (
	resultPath        = "/tmp/result.json"
	staffLogPath      = "/tmp/staff.log"
	contestantLogPath = "/tmp/contestant.log"
)

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

func joinN(messages []string, n int) string {
	if len(messages) > n {
		return strings.Join(messages[:n], ",\n")
	}
	return strings.Join(messages, ",\n")
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
	for _, name := range []string{resultPath, staffLogPath, contestantLogPath} {
		os.Remove(name)
	}

	benchOptions := []string{
		"run",
		"--nameserver", target,
		"--staff-log-path", staffLogPath,
		"--contestant-log-path", contestantLogPath,
		"--result-path", resultPath,
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
	log.Println("===== options =====")
	for _, opt := range benchOptions {
		log.Printf("%s\n", opt)
	}

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
		if err := cmd.Run(); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		NotifyWorkerErr(job, ctx.Err(), stdout.String(), stderr.String(), "ベンチマーカーの実行がタイムアウトしました (StatusFailed)")
		status = StatusTimeout
	case err, ok := <-errCh:
		if ok && err != nil {
			NotifyWorkerErr(job, err, stdout.String(), stderr.String(), "ベンチマーカーの実行エラーが発生 (StatusFailed)")
			status = StatusFailed
		}
	}

	log.Println(stdout.String())

	var msgs []string
	// stdout
	b, err := os.ReadFile(contestantLogPath)
	if err != nil {
		return &Result{
			ID:         job.ID,
			Stdout:     stdout.String(),
			Stderr:     stderr.String(),
			Reason:     err.Error(),
			IsPassed:   false,
			Score:      0,
			Status:     status,
			FinishedAt: time.Now(),
		}, nil
	}
	contestantLog := strings.Split(string(b), "\n")

	// result
	var benchResult *BenchResult
	b, err = os.ReadFile(resultPath)
	if err != nil {
		return &Result{
			ID:         job.ID,
			Stdout:     stdout.String(),
			Stderr:     stderr.String(),
			Reason:     err.Error(),
			IsPassed:   false,
			Score:      0,
			Status:     status,
			FinishedAt: time.Now(),
		}, nil
	}
	if err := json.NewDecoder(bytes.NewBuffer(b)).Decode(&benchResult); err != nil {
		return &Result{
			ID:         job.ID,
			Stdout:     stdout.String(),
			Stderr:     stderr.String(),
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
		return &Result{
			ID:            job.ID,
			Stdout:        stdout.String(),
			Stderr:        stderr.String(),
			Reason:        joinN(msgs, messageLimit),
			IsPassed:      benchResult.Pass,
			Score:         benchResult.Score,
			ResolvedCount: benchResult.ResolvedCount,
			Status:        status,
			FinishedAt:    time.Now(),
		}, nil
	} else {
		return &Result{
			ID:         job.ID,
			Stdout:     stdout.String(),
			Stderr:     stderr.String(),
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
			Value:       "AKIAWFVKEZX5GDVVMWF7",
			Destination: &accessKey,
			EnvVar:      "SUPERVISOR_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:        "secret-access-key",
			Value:       "WXrWx7UWZIN85IzCoK8dGvHFivU+jZvcUhWdi21i",
			Destination: &secretAccessKey,
			EnvVar:      "SUPERVISOR_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:        "slack-webhook-url",
			Value:       "https://hooks.slack.com/services/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",
			Destination: &slackWebhookURL,
			EnvVar:      "SUPERVISOR_SLACK_WEBHOOK_URL",
		},
		cli.IntFlag{
			Name:        "message-limit",
			Value:       2000,
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

		var (
			portal *Portal
			signer ssh.Signer
		)
		if production {
			accessKey = "AKIAWFVKEZX5AUP2AK6O"
			secretAccessKey = "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"
			log.Println("Running on production")
			azName, err := fetchAZName(ctx)
			if err != nil {
				return cli.NewExitError(err, 1)
			}

			portal, err = NewPortal(
				azName,
				Production,
				accessKey,
				secretAccessKey,
			)
			if err != nil {
				return cli.NewExitError(err, 1)
			}

			privateKey, err := os.ReadFile("/home/benchuser/cmd/bench/id_ed25519")
			if err != nil {
				log.Printf("privateKey error = %s\n", err)
				return err
			}
			signer, err = ssh.ParsePrivateKey(privateKey)
			if err != nil {
				log.Printf("signer error = %s\n", err)
				return err
			}

		} else {
			log.Println("Running on development")
			azName, err := fetchAZName(ctx)
			if err != nil {
				return cli.NewExitError(err, 1)
			}

			// az1, az2, az4
			portal, err = NewPortal(
				azName,
				Develop,
				accessKey,
				secretAccessKey,
			)
			if err != nil {
				return cli.NewExitError(err, 1)
			}

			privateKey, err := os.ReadFile("/home/benchuser/cmd/bench/id_ed25519")
			if err != nil {
				log.Printf("privateKey error = %s\n", err)
				return err
			}
			signer, err = ssh.ParsePrivateKey(privateKey)
			if err != nil {
				log.Printf("signer error = %s\n", err)
				return err
			}
		}
		jobCh := portal.StartReceiveJob(ctx)

		log.Printf("AccessKey = ...%s\n", accessKey[len(accessKey)-1-3:len(accessKey)-1])
		for {
			select {
			case <-ctx.Done():
				return cli.NewExitError(ctx.Err(), 1)
			case job := <-jobCh:
				log.Printf("job = %+v\n", job)

				if job.Action == "reboot" {
					var errs []string
					for _, server := range job.Servers {
						if err := reboot(server, signer); err != nil {
							errs = append(errs, err.Error())
						}
					}
					if len(errs) > 0 {
						NotifyWorkerErr(job, nil, "", "", strings.Join(errs, ","))
					}
				} else {
					if err := portal.SendResult(ctx, job, NewRunningResult(job.ID)); err != nil {
						NotifyWorkerErr(job, err, "", "", "ベンチマーカーの実行に失敗。すぐに調査してください。supervisorの処理は継続します")
					}

					result, err := execBench(ctx, job)
					if err != nil {
						NotifyWorkerErr(job, err, "", "", "ベンチマーカーの実行に失敗。すぐに調査してください。supervisorの処理は継続します")
					}
					log.Printf("finishedAt = %s\n", result.FinishedAt.String())

					sendResultStartAt := time.Now()
					log.Printf("sendResultStartAt = %s\n", sendResultStartAt.String())
					if err := portal.SendResult(ctx, job, result); err != nil {
						NotifyWorkerErr(job, err, "", "", "ベンチマーカーの結果送信に失敗。すぐに調査してください。supervisorの処理は継続します")
					}
					log.Printf("sendResultFinishedAt = %s\n", time.Since(sendResultStartAt).String())

					os.Remove(staffLogPath)
					os.Remove(contestantLogPath)
					os.Remove(resultPath)
				}
			}
		}
	},
}
