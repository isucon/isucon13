package scenario

import (
	"context"
	"log"
	"time"

	"github.com/isucon/isucon13/bench/isupipe"
)

func Pretest(ctx context.Context, client *isupipe.Client) error {
	user, err := client.PostUser(ctx, &isupipe.PostUserRequest{
		Name:        "test",
		DisplayName: "test",
		Description: "blah blah blah",
		Password:    "s3cr3t",
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}

	log.Printf("try to login...")
	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: user.Name,
		Password: "s3cr3t",
	}); err != nil {
		return err
	}

	log.Printf("try to get user...")
	if err := client.GetUser(ctx, user.ID /* user id */); err != nil {
		return err
	}

	log.Printf("try to get users...")
	if _, err := client.GetUsers(ctx); err != nil {
		return err
	}

	log.Printf("try to get streamer theme...")
	// FIXME

	// if err := client.GetStreamerTheme(ctx, user.ID /* user id */); err != nil {
	// return err
	// }

	log.Printf("try to get tags...")
	if _, err := client.GetTags(ctx); err != nil {
		return err
	}

	log.Printf("try to reserve livestream...")
	livestream, err := client.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
		Tags:        []int{1},
		Title:       "test",
		Description: "test",
		StartAt:     time.Date(2024, 07, 12, 1, 0, 0, 0, time.Local).Unix(),
		EndAt:       time.Date(2024, 07, 12, 2, 0, 0, 0, time.Local).Unix(),
	})
	if err != nil {
		return err
	}

	log.Printf("try to get livecomment reports...")
	if _, err = client.GetLivecommentReports(ctx, livestream.Id); err != nil {
		return err
	}
	log.Printf("try to get livestream...")
	if err = client.GetLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	log.Printf("try to enter livestream...")
	if err := client.EnterLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	log.Printf("try to post livecomment...")
	livecomment, err := client.PostLivecomment(ctx, livestream.Id, &isupipe.PostLivecommentRequest{
		Comment: "test",
		Tip:     3,
	})
	if err != nil {
		return err
	}

	log.Printf("try to get livecomments...")
	if _, err := client.GetLivecomments(ctx, livestream.Id /* livestream id*/); err != nil {
		return err
	}

	log.Printf("try to report livecomment...")
	if err := client.ReportLivecomment(ctx, livestream.Id, livecomment.Id); err != nil {
		return err
	}

	log.Printf("try to post reaction...")
	if _, err := client.PostReaction(ctx, livestream.Id /* livestream id*/, &isupipe.PostReactionRequest{
		EmojiName: ":chair:",
	}); err != nil {
		return err
	}

	log.Printf("try to get reactions...")
	if _, err := client.GetReactions(ctx, livestream.Id /* livestream id*/); err != nil {
		return err
	}

	log.Printf("try to get livestreams by tag...")
	if err := client.GetLivestreamsByTag(ctx, "chair" /* tag name */); err != nil {
		return err
	}

	log.Printf("try to leave livestream...")
	if err := client.LeaveLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	return nil
}
