package main

import (
	"context"
	"log"

	"github.com/isucon/isucon13/bench/isupipe"
	"github.com/isucon/isucon13/bench/scenario"
)

func main() {
	ctx := context.Background()

	client, err := isupipe.NewClient()
	if err != nil {
		log.Fatalln(err)
	}

	if err := scenario.Pretest(ctx, client); err != nil {
		log.Fatalln(err)
	}
}
