package scenario

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
)

const testUserRawPassword = "s3cr3t"

func Pretest(ctx context.Context, client *isupipe.Client) error {
	// FIXME: 処理前、paymentが0円になってることをチェック
	// FIXME: 処理後、paymentが指定金額になっていることをチェック

	// FIXME: 処理前、統計情報がすべて0になっていることをチェック
	// FIXME: いくつかの処理後、統計情報がピタリ一致することをチェック
	//        (処理数、処理データにランダム性をもたせる)

	user, err := client.Register(ctx, &isupipe.RegisterRequest{
		Name:        "test",
		DisplayName: "test",
		Description: "blah blah blah",
		Password:    testUserRawPassword,
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	})
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		UserName: user.Name,
		Password: testUserRawPassword,
	}); err != nil {
		return err
	}

	if _, err := client.GetMe(ctx); err != nil {
		return err
	}

	if err := client.GetUser(ctx, user.Name); err != nil {
		return err
	}

	if _, err := client.GetUsers(ctx); err != nil {
		return err
	}

	if _, err := client.GetStreamerTheme(ctx, user); err != nil {
		return err
	}

	tagResponse, err := client.GetTags(ctx)
	if err != nil {
		return err
	}
	if len(tagResponse.Tags) != scheduler.GetTagPoolLength() {
		return fmt.Errorf("初期データのタグが正常に登録されていません: want=%d, but got=%d", scheduler.GetTagPoolLength(), len(tagResponse.Tags))
	}

	var (
		tagCount    = rand.Intn(5)
		tagStartIdx = rand.Intn(len(tagResponse.Tags))
		tagEndIdx   = min(tagStartIdx+tagCount, len(tagResponse.Tags))
	)
	var tags []int
	for _, tag := range tagResponse.Tags[tagStartIdx:tagEndIdx] {
		tags = append(tags, tag.Id)
	}

	// FIXME: 枠数を超えて予約した場合にエラーになるか

	reservation, err := scheduler.ReservationSched.GetColdReservation()
	if err != nil {
		return err
	}
	livestream, err := client.ReserveLivestream(ctx, &isupipe.ReserveLivestreamRequest{
		// FIXME: webapp側でタグの採番がおかしく、エラーが出るので一時的に無効化
		Tags:        tags,
		Title:       reservation.Title,
		Description: reservation.Description,
		StartAt:     reservation.StartAt,
		EndAt:       reservation.EndAt,
	})
	if err != nil {
		scheduler.ReservationSched.AbortReservation(reservation)
		return err
	}
	scheduler.ReservationSched.CommitReservation(reservation)

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
		EmojiName: "chair",
	}); err != nil {
		return err
	}

	if _, err := client.GetReactions(ctx, livestream.Id /* livestream id*/); err != nil {
		return err
	}

	if err := client.GetLivestreamsByTag(ctx, "椅子" /* tag name */); err != nil {
		return err
	}

	if _, err := client.GetUserStatistics(ctx, user.Name); err != nil {
		return err
	}

	if _, err := client.GetLivestreamStatistics(ctx, livestream.Id); err != nil {
		return err
	}

	if err := client.LeaveLivestream(ctx, livestream.Id); err != nil {
		return err
	}

	if err := assertBadLogin(ctx, client, user); err != nil {
		return err
	}
	if err := assertPipeUserRegistration(ctx, client); err != nil {
		return err
	}
	if err := assertUserUniqueConstraint(ctx, client); err != nil {
		return err
	}

	return nil
}

func assertPipeUserRegistration(ctx context.Context, client *isupipe.Client) error {
	// pipeユーザが弾かれることを確認
	pipeReq := isupipe.RegisterRequest{
		Name:        "pipe",
		DisplayName: "pipe",
		Description: "blah blah blah",
		Password:    "pipe",
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	}
	if _, err := client.Register(ctx, &pipeReq, isupipe.WithStatusCode(http.StatusBadRequest)); err != nil {
		return fmt.Errorf("'pipe'ユーザの作成は拒否されなければなりません: %w", err)
	}

	return nil
}

func assertUserUniqueConstraint(ctx context.Context, client *isupipe.Client) error {
	testDupReq := isupipe.RegisterRequest{
		Name:        "aaa",
		DisplayName: "hoge",
		Description: "lorem ipsum",
		Password:    "hogefugaaaa",
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	}
	if _, err := client.Register(ctx, &testDupReq); err != nil {
		return err
	}

	if _, err := client.Register(ctx, &testDupReq, isupipe.WithStatusCode(http.StatusInternalServerError)); err != nil {
		return fmt.Errorf("重複したユーザ名を含むリクエストはエラーを返さなければなりません: %w", err)
	}

	return nil
}

func assertBadLogin(ctx context.Context, client *isupipe.Client, user *isupipe.User) error {
	// 存在しないユーザでログインされた場合はエラー
	unknownUserReq := isupipe.LoginRequest{
		UserName: "unknownUser4328904823",
		Password: "unknownUser",
	}

	if err := client.Login(ctx, &unknownUserReq, isupipe.WithStatusCode(http.StatusUnauthorized)); err != nil {
		return bencherror.NewViolationError(err, "データベースに存在しないユーザからのログインは無効です")
	}

	// パスワードが間違っている場合はエラー
	wrongPasswordReq := isupipe.LoginRequest{
		UserName: user.Name,
		Password: "wrongPassword",
	}
	if err := client.Login(ctx, &wrongPasswordReq, isupipe.WithStatusCode(http.StatusUnauthorized)); err != nil {
		return bencherror.NewViolationError(err, "パスワードが間違っているログインは無効です")
	}

	return nil
}
