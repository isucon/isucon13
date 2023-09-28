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

	log.Printf("try to get user theme...")
	if err := client.GetUserTheme(ctx, user.ID /* user id */); err != nil {
		return err
	}

	log.Printf("try to get tags...")
	if err := client.GetTags(ctx); err != nil {
		return err
	}

	log.Printf("try to reserve livestream...")
	livestream, err := client.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
		Title:         "test",
		Description:   "test",
		PrivacyStatus: "public",
		StartAt:       time.Now().Unix(),
		EndAt:         time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	log.Printf("try to enter livestream...")
	if err := client.EnterLivestream(ctx, livestream.ID); err != nil {
		return err
	}

	log.Printf("try to post superchat...")
	superchat, err := client.PostSuperchat(ctx, livestream.ID, &isupipe.PostSuperchatRequest{
		Comment: "test",
		Tip:     3,
	})
	if err != nil {
		return err
	}

	log.Printf("try to get superchats...")
	if _, err := client.GetSuperchats(ctx, livestream.ID /* livestream id*/); err != nil {
		return err
	}

	log.Printf("try to report superchat...")
	if err := client.ReportSuperchat(ctx, superchat.ID); err != nil {
		return err
	}

	log.Printf("try to post reaction...")
	if _, err := client.PostReaction(ctx, livestream.ID /* livestream id*/, &isupipe.PostReactionRequest{
		EmojiName: ":chair:",
	}); err != nil {
		return err
	}

	log.Printf("try to get reactions...")
	if _, err := client.GetReactions(ctx, livestream.ID /* livestream id*/); err != nil {
		return err
	}

	log.Printf("try to get livestreams by tag...")
	if err := client.GetLivestreamsByTag(ctx, "chair" /* tag name */); err != nil {
		return err
	}

	log.Printf("try to leave jlivestream...")
	if err := client.LeaveLivestream(ctx, livestream.ID); err != nil {
		return err
	}

	return nil
}
