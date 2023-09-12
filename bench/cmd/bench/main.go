package main

import (
	"context"
	"log"
	"time"

	"github.com/isucon/isucon13/bench/isupipe"
)

func main() {
	ctx := context.Background()

	client, err := isupipe.NewClient()
	if err != nil {
		log.Fatalln(err)
	}

	if err := client.PostUser(ctx, &isupipe.PostUserRequest{
		Name:        "test",
		DisplayName: "test",
		Description: "blah blah blah",
		Password:    "s3cr3t",
	}); err != nil {
		log.Fatalln(err)
	}
	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: "test",
		Password: "s3cr3t",
	}); err != nil {
		log.Fatalln(err)
	}

	//
	if err := client.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
		Title:         "test",
		Description:   "test",
		PrivacyStatus: "public",
		StartAt:       time.Now().Unix(),
		EndAt:         time.Now().Unix(),
	}); err != nil {
		log.Fatalln(err)
	}

	if err := client.PostSuperchat(ctx, 1, &isupipe.PostSuperchatRequest{
		Comment: "test",
		Tip:     3,
	}); err != nil {
		log.Fatalln(err)
	}
}
