package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const queueBaseUrl = "https://sqs.ap-northeast-1.amazonaws.com/424484851194"

type Environment string

var (
	Develop    Environment = "develop"
	Production Environment = "prod"
)

type Result struct {
	ID         int       `json:"id"`
	Status     string    `json:"status"`
	Score      int64     `json:"score,omitempty"`
	IsPassed   bool      `json:"is_passed,omitempty"`
	Reason     string    `json:"reason,omitempty"`
	Stdout     string    `json:"stdout,omitempty"`
	Stderr     string    `json:"stderr,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

func NewAbortResult(id int) *Result {
	return &Result{
		ID:     id,
		Status: "aborted",
	}
}

type Job struct {
	ID       int      `json:"id"`
	Team     int      `json:"team"`
	TargetIP string   `json:"target_ip"`
	Servers  []string `json:"servers"`
}

type Portal struct {
	client *sqs.SQS

	sendQueueUrl string
	recvQueueUrl string
}

func NewPortal(
	azID string,
	environment Environment,
	accessKey, secretAccessKey string,
) (*Portal, error) {
	if !strings.HasPrefix(azID, "apne1-az") {
		return nil, fmt.Errorf("invalid availability zone: %s", azID)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("ap-northeast-1"),
		Credentials: credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
	}))
	client := sqs.New(sess)

	// NOTE: ポータルへの送信キューはAZ分散しない
	sendQueueName := fmt.Sprintf("%s-job-queue.fifo", environment)
	sendQueueUrl, err := url.JoinPath(queueBaseUrl, sendQueueName)
	if err != nil {
		return nil, err
	}
	// NOTE: ポータルからの受信キューはAZ分散する
	recvQueueName := fmt.Sprintf("%s-job-queue-%s.fifo", environment, azID)
	recvQueueUrl, err := url.JoinPath(queueBaseUrl, recvQueueName)
	if err != nil {
		return nil, err
	}
	return &Portal{
		client:       client,
		sendQueueUrl: sendQueueUrl,
		recvQueueUrl: recvQueueUrl,
	}, nil
}

func (s *Portal) SendResult(ctx context.Context, job *Job, result *Result) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if _, err := s.client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.sendQueueUrl),
		MessageBody: aws.String(string(b)),
	}); err != nil {
		return err
	}

	return nil
}

func (s *Portal) StartReceiveJob(ctx context.Context) <-chan *Job {
	ch := make(chan *Job)
	go func() {
		defer close(ch)
	loop:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				resp, err := s.client.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
					QueueUrl:            aws.String(s.recvQueueUrl),
					MaxNumberOfMessages: aws.Int64(1),
					WaitTimeSeconds:     aws.Int64(1),
				})
				if err != nil {
					continue loop
				}
				if len(resp.Messages) == 0 {
					continue loop
				}

				msg := resp.Messages[0]

				var job *Job
				if err := json.NewDecoder(strings.NewReader(*msg.Body)).Decode(&job); err != nil {
					log.Printf("failed to decode json: %s\n", err.Error())
					continue loop
				}

				// NOTE: ジョブを取れたらすぐに削除 (可視性タイムアウト)
				if _, err := s.client.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      aws.String(s.recvQueueUrl),
					ReceiptHandle: msg.ReceiptHandle,
				}); err != nil {
					continue loop
				}

				ch <- job
			}
		}
	}()
	return ch
}
