package scenario

import (
	"context"
	"log"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
)

const testUserRawPassword = "s3cr3t"

func Pretest(ctx context.Context, client *isupipe.Client) error {
	// FIXME: 処理前、paymentが0円になってることをチェック
	// FIXME: 処理後、paymentが指定金額になっていることをチェック

	user, err := postUserPretest(ctx, client)
	if err != nil {
		return err
	}

	if err := loginPretest(ctx, client, user); err != nil {
		return err
	}

	log.Printf("try to get user(me)...")
	if err := client.GetUserSession(ctx); err != nil {
		return err
	}

	log.Printf("try to get user...")
	if err := client.GetUser(ctx, user.Id /* user id */); err != nil {
		return err
	}

	if _, err := client.GetUsers(ctx); err != nil {
		log.Printf("failed to get users %s", err.Error())
		return err
	}

	// FIXME

	// if err := client.GetStreamerTheme(ctx, user.Id /* user id */); err != nil {
	// return err
	// }

	if err := getTagsPretest(ctx, client); err != nil {
		return err
	}
	if _, err := client.GetTags(ctx); err != nil {
		return err
	}

	reservation, err := scheduler.Phase2ReservationScheduler.GetHotShortReservation()
	if err != nil {
		return err
	}
	livestream, err := client.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
		Tags:        []int{},
		Title:       reservation.Title,
		Description: reservation.Description,
		StartAt:     reservation.StartAt,
		EndAt:       reservation.EndAt,
	})
	if err != nil {
		scheduler.Phase2ReservationScheduler.AbortReservation(reservation)
		return err
	}
	scheduler.Phase2ReservationScheduler.CommitReservation(reservation)

	if _, err = client.GetLivecommentReports(ctx, livestream.Id); err != nil {
		return err
	}
	if err = client.GetLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	if err := client.EnterLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	livecomment, err := client.PostLivecomment(ctx, livestream.Id, &isupipe.PostLivecommentRequest{
		Comment: "test",
		Tip:     3,
	})
	if err != nil {
		return err
	}

	if _, err := client.GetLivecomments(ctx, livestream.Id /* livestream id*/); err != nil {
		return err
	}

	if err := client.ReportLivecomment(ctx, livestream.Id, livecomment.Id); err != nil {
		return err
	}

	if _, err := client.PostReaction(ctx, livestream.Id /* livestream id*/, &isupipe.PostReactionRequest{
		EmojiName: ":chair:",
	}); err != nil {
		return err
	}

	if _, err := client.GetReactions(ctx, livestream.Id /* livestream id*/); err != nil {
		return err
	}

	if err := client.GetLivestreamsByTag(ctx, "椅子" /* tag name */); err != nil {
		return err
	}

	if _, err := client.GetUserStatistics(ctx, user.Id); err != nil {
		return err
	}

	if _, err := client.GetLivestreamStatistics(ctx, livestream.Id); err != nil {
		return err
	}

	if err := client.LeaveLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	return nil
}

func postUserPretest(ctx context.Context, client *isupipe.Client) (*isupipe.User, error) {
	// pipeユーザが弾かれることを確認
	pipeReq := isupipe.PostUserRequest{
		Name:        "pipe",
		DisplayName: "pipe",
		Description: "blah blah blah",
		Password:    "pipe",
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	}
	if _, err := client.PostUser(ctx, &pipeReq, isupipe.WithStatusCode(http.StatusBadRequest), isupipe.DecodeBody(false)); err != nil {
		return nil, bencherror.NewViolationError(err, "'pipe'ユーザの作成は拒否されなければなりません")
	}

	// 正常系検証
	testUserReq := isupipe.PostUserRequest{
		Name:        "test",
		DisplayName: "test",
		Description: "blah blah blah",
		Password:    testUserRawPassword,
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	}
	user, err := client.PostUser(ctx, &testUserReq)
	if err != nil {
		return nil, err
	}

	// usernameが重複する場合は作成に失敗すること
	if _, err := client.PostUser(ctx, &testUserReq, isupipe.WithStatusCode(http.StatusInternalServerError), isupipe.DecodeBody(false)); err != nil {
		return nil, bencherror.NewViolationError(err, "重複したユーザ名を含むリクエストはエラーを返さなければなりません")
	}

	return user, nil
}

func loginPretest(ctx context.Context, client *isupipe.Client, user *isupipe.User) error {
	// 正常系検査
	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: user.Name,
		Password: testUserRawPassword,
	}); err != nil {
		return err
	}

	// 存在しないユーザでログインされた場合はエラー
	unknownUserReq := isupipe.LoginRequest{
		UserName: "unknownUser4328904823",
		Password: "unknownUser",
	}

	if err := client.Login(ctx, &unknownUserReq, isupipe.WithStatusCode(http.StatusUnauthorized), isupipe.DecodeBody(false)); err != nil {
		return bencherror.NewViolationError(err, "データベースに存在しないユーザからのログインは無効です")
	}

	// パスワードが間違っている場合はエラー
	wrongPasswordReq := isupipe.LoginRequest{
		UserName: user.Name,
		Password: "wrongPassword",
	}
	if err := client.Login(ctx, &wrongPasswordReq, isupipe.WithStatusCode(http.StatusUnauthorized), isupipe.DecodeBody(false)); err != nil {
		return bencherror.NewViolationError(err, "パスワードが間違っているログインは無効です")
	}

	return nil
}

func getTagsPretest(ctx context.Context, client *isupipe.Client) error {
	const preDefinedTagCount = 103
	tags, err := client.GetTags(ctx)
	if err != nil {
		return err
	}

	if len(tags.Tags) != preDefinedTagCount {
		return bencherror.NewViolationError(nil, "事前定義されたタグの数が足りません")
	}

	return nil
}
