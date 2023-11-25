package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/isucon/isucon13/bench/internal/config"
)

func UploadFinalcheckResult(bucketName string, jobID, team int) error {
	fd, err := os.Open(config.FinalcheckPath)
	if err != nil {
		return err
	}
	defer fd.Close()

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String("ap-northeast-1"),
		Credentials: credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
	}))
	client := s3.New(sess)

	key := fmt.Sprintf("team-%d-job-%d-finalcheck.json", team, jobID)
	if _, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   fd,
	}); err != nil {
		return err
	}

	return nil
}
