package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli"
)

// FIXME: SQSのメッセージサイズが最大で256KBなので、200KB程度までで打ち切るように

const (
	resultPath        = "/tmp/result.json"
	staffLogPath      = "/tmp/staff.log"
	contestantLogPath = "/tmp/contestant.log"
)

var (
	sendQueueUrl, recvQueueUrl string
	accessKey, secretAccessKey string
	slackWebhookURL            string
	messageLimit               int
)

const (
	MsgTimeout = "ベンチマーク処理がタイムアウトしました"
	MsgFail    = "運営に連絡してください"
)

const (
	StatusSuccess = "done"
	StatusFailed  = "aborted"
	StatusTimeout = "aborted"
)

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
		if err := os.Remove(name); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	benchOptions := []string{
		"run",
		"--nameserver", target,
		"--staff-log-path", staffLogPath,
		"--contestant-log-path", contestantLogPath,
		"--result-path", resultPath,
	}
	if enableSSL {
		benchOptions = append(benchOptions, "--enable-ssl")
		benchOptions = append(benchOptions, "--target", "https://pipe.u.isucon.dev:443")
	} else {
		// NOTE: 開発環境(Docker)
		benchOptions = append(benchOptions, "--dns-port", strconv.Itoa(1053))
		benchOptions = append(benchOptions, "--target", "http://pipe.u.isucon.dev:8080")
	}
	log.Println("===== options =====")
	for _, opt := range benchOptions {
		log.Printf("%s\n", opt)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, executablePath, benchOptions...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = 1 * time.Minute

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
		NotifyWorkerErr(job, ctx.Err(), stdout.String(), stderr.String(), "ベンチマーカーの実行中context timeoutが発生 (StatusFailed)")
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
			ID:       job.ID,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Reason:   err.Error(),
			IsPassed: false,
			Score:    0,
			Status:   status,
		}, nil
	}
	contestantLog := strings.Split(string(b), "\n")

	// result
	var benchResult *BenchResult
	b, err = os.ReadFile(resultPath)
	if err != nil {
		return &Result{
			ID:       job.ID,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Reason:   err.Error(),
			IsPassed: false,
			Score:    0,
			Status:   status,
		}, nil
	}
	if err := json.NewDecoder(bytes.NewBuffer(b)).Decode(&benchResult); err != nil {
		return &Result{
			ID:       job.ID,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Reason:   err.Error(),
			IsPassed: false,
			Score:    0,
			Status:   status,
		}, nil
	}

	msgs = contestantLog
	msgs = append(msgs, benchResult.Messages...)

	if status == StatusSuccess {
		return &Result{
			ID:       job.ID,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Reason:   joinN(msgs, messageLimit),
			IsPassed: benchResult.Pass,
			Score:    benchResult.Score,
			Status:   status,
		}, nil
	} else {
		return &Result{
			ID:       job.ID,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Reason:   "ベンチマーク失敗",
			IsPassed: false,
			Score:    0,
			Status:   status,
		}, nil
	}
}

var supervise = cli.Command{
	Name:  "supervise",
	Usage: "supervisor実行",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:        "send-queue-url",
			Value:       " https://sqs.ap-northeast-1.amazonaws.com/424484851194/develop-job-result",
			Destination: &sendQueueUrl,
			EnvVar:      "SUPERVISOR_SEND_QUEUE_URL",
		},
		cli.StringFlag{
			Name:        "recv-queue-url",
			Value:       "https://sqs.ap-northeast-1.amazonaws.com/424484851194/develop-job-queue.fifo",
			Destination: &recvQueueUrl,
			EnvVar:      "SUPERVISOR_RECV_QUEUE_URL",
		},
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
			Value:       3000,
			Destination: &messageLimit,
			EnvVar:      "SUPERVISOR_MESSAGE_LIMIT",
		},
		cli.BoolFlag{
			Name:        "enable-ssl",
			Destination: &enableSSL,
			EnvVar:      "SUPERVISOR_ENABLE_SSL",
		},
	},
	Action: func(cliCtx *cli.Context) error {
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP)
		defer cancel()

		portal := NewPortal(
			sendQueueUrl,
			recvQueueUrl,
			accessKey,
			secretAccessKey,
		)
		jobCh := portal.StartReceiveJob(ctx)

		for {
			select {
			case <-ctx.Done():
				return cli.NewExitError(ctx.Err(), 1)
			case job := <-jobCh:
				log.Printf("job = %+v\n", job)

				result, err := execBench(ctx, job)
				if err != nil {
					NotifyWorkerErr(job, err, "", "", "ベンチマーカーの実行に失敗。すぐに調査してください。supervisorの処理は継続します")
				}

				if err := portal.SendResult(ctx, job, result); err != nil {
					NotifyWorkerErr(job, err, "", "", "ベンチマーカーの結果送信に失敗。すぐに調査してください。supervisorの処理は継続します")
				}

				os.Remove(staffLogPath)
				os.Remove(contestantLogPath)
				os.Remove(resultPath)
			}
		}
	},
}
