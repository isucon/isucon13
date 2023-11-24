package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/nlopes/slack"
)

func notifyErr(title string, err error, attachments []slack.Attachment) error {
	if len(slackWebhookURL) == 0 {
		return nil
	}

	if err != nil {
		attachments = append(attachments, slack.Attachment{
			Color: "danger",
			Title: "エラー情報",
			Text:  err.Error(),
		})
	}

	var (
		retryCnt      = 10
		retryInterval = 1 * time.Second
	)
	postRetrier := retrier.New(retrier.ConstantBackoff(retryCnt, retryInterval), nil)
	return postRetrier.Run(func() error {
		return slack.PostWebhook(slackWebhookURL, &slack.WebhookMessage{
			// Text:        fmt.Sprintf("<!channel> %s", title),
			Text:        title,
			Attachments: attachments,
		})
	})
}

func NotifyWorkerErr(job *Job, err error, stdout, stderr string, msg string, args ...interface{}) error {
	hostname, hostnameErr := os.Hostname()
	if hostnameErr != nil {
		return hostnameErr
	}

	attachments := []slack.Attachment{
		slack.Attachment{
			Color: "danger",
			Title: "補足情報",
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: "ホスト名",
					Value: hostname,
					Short: true,
				},
				slack.AttachmentField{
					Title: "ジョブID",
					Value: fmt.Sprintf("%d", job.ID),
					Short: true,
				},
				slack.AttachmentField{
					Title: "チームID",
					Value: fmt.Sprintf("%d", job.Team),
					Short: true,
				},
				slack.AttachmentField{
					Title: "AZ名",
					Value: AZName,
					Short: true,
				},
				slack.AttachmentField{
					Title: "メッセージ",
					Value: fmt.Sprintf(msg, args...),
					Short: true,
				},
			},
		},
	}

	if len(stdout) > 0 {
		attachments = append(attachments, slack.Attachment{
			Color: "danger",
			Title: "標準出力",
			Text:  string(stdout),
		})
	}

	if len(stderr) > 0 {
		attachments = append(attachments, slack.Attachment{
			Color: "danger",
			Title: "標準エラー出力",
			Text:  string(stderr),
		})
	}

	return notifyErr("エラー発生", err, attachments)
}
