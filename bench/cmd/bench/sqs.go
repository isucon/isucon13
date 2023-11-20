package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
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

	receiptHandle *string
}

type Portal struct {
	client *sqs.SQS

	sendQueueUrl string
	recvQueueUrl string
}

func NewPortal(
	sendQueueUrl string,
	recvQueueUrl string,
	accessKey, secretAccessKey string,
) *Portal {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("ap-northeast-1"),
		Credentials: credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
	}))
	client := sqs.New(sess)
	return &Portal{
		client:       client,
		sendQueueUrl: sendQueueUrl,
		recvQueueUrl: recvQueueUrl,
	}
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

	if _, err := s.client.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.recvQueueUrl),
		ReceiptHandle: job.receiptHandle,
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
				job.receiptHandle = msg.ReceiptHandle

				ch <- job
			}
		}
	}()
	return ch
}
