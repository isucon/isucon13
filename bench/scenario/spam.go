package scenario

import (
	"context"
	"time"

	"github.com/isucon/isucon13/bench/isupipe"
)

func SpamScenario(ctx context.Context, client *isupipe.Client) error {
	viewer, err := client.PostUser(ctx, &isupipe.PostUserRequest{
		Name:        "spamscenario",
		DisplayName: "spamscenario",
		Description: "SpanScenario",
		Password:    "spamspamspam",
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: viewer.Name,
		Password: "spamspamspam",
	}); err != nil {
		return err
	}

	livestream, err := client.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
		Tags:        []int{1},
		Title:       "SpamScenario",
		Description: "spam",
		StartAt:     time.Date(2024, 07, 10, 1, 0, 0, 0, time.Local).Unix(),
		EndAt:       time.Date(2024, 07, 10, 2, 0, 0, 0, time.Local).Unix(),
	})
	if err != nil {
		return err
	}

	// シードデータの配信予約に対し、スパム投稿
	superchat, err := client.PostSuperchat(ctx, livestream.Id, &isupipe.PostSuperchatRequest{
		Comment: "this is spam",
		Tip:     0,
	})
	if err != nil {
		return err
	}

	// 特定スパチャをスパム報告
	if err := client.ReportSuperchat(ctx, livestream.Id, superchat.Id); err != nil {
		return err
	}

	if err := client.Moderate(ctx, livestream.Id, "spam"); err != nil {
		return err
	}
	// FIXME: スパムで弾かれるはずの投げたメッセージを再度投げ直す -> 弾かれることを確認

	return nil
}
