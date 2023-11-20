package scenario

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/bencherror"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/isupipe"
)

// 異常系

func assertPipeUserRegistration(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	// pipeユーザが弾かれることを確認
	if _, err := client.Register(ctx, &isupipe.RegisterRequest{
		Name:        "pipe",
		DisplayName: "pipe",
		Description: "blah blah blah",
		Password:    "pipe",
		Theme: isupipe.Theme{
			DarkMode: true,
		},
	}, isupipe.WithStatusCode(http.StatusBadRequest)); err != nil {
		return fmt.Errorf("'pipe'ユーザの作成は拒否されなければなりません: %w", err)
	}

	return nil
}

func assertBadLogin(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	// 存在しないユーザでログインされた場合はエラー
	client1, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(3*time.Second),
	)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	unknownUserReq := isupipe.LoginRequest{
		Username: "unknownUser4328904823",
		Password: "unknownUser",
	}

	if err := client1.Login(ctx, &unknownUserReq, isupipe.WithStatusCode(http.StatusUnauthorized)); err != nil {
		return bencherror.NewViolationError(err, "データベースに存在しないユーザからのログインは無効です")
	}

	// パスワードが間違っている場合はエラー
	client2, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(3*time.Second),
	)
	if err != nil {
		return bencherror.NewInternalError(err)
	}
	wrongPasswordReq := isupipe.LoginRequest{
		Username: "test001",
		Password: "wrongPassword",
	}
	if err := client2.Login(ctx, &wrongPasswordReq, isupipe.WithStatusCode(http.StatusUnauthorized)); err != nil {
		return bencherror.NewViolationError(err, "パスワードが間違っているログインは無効です")
	}

	return nil
}

func assertUserUniqueConstraint(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

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

func assertReserveOverflowPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	// NumSlotを超えて予約しようとするとエラーになる
	var overflow bool
	for idx := 0; idx < config.NumSlots; idx++ {
		overflowClient, err := isupipe.NewCustomResolverClient(
			dnsResolver,
			agent.WithTimeout(config.PretestTimeout),
		)
		if err != nil {
			return err
		}

		name := fmt.Sprintf("overflow%d", idx)
		overflowUser, err := overflowClient.Register(ctx, &isupipe.RegisterRequest{
			Name:        name,
			DisplayName: name,
			Description: name,
			Password:    "test",
			Theme: isupipe.Theme{
				DarkMode: true,
			},
		})
		if err != nil {
			return err
		}
		if err := overflowClient.Login(ctx, &isupipe.LoginRequest{
			Username: overflowUser.Name,
			Password: "test",
		}); err != nil {
			return err
		}

		var (
			startAt = time.Date(2024, 4, 1, 0, 0, 0, 0, time.Local)
			endAt   = time.Date(2024, 4, 1, 1, 0, 0, 0, time.Local)
		)
		_, err = overflowClient.ReserveLivestream(ctx, overflowUser.Name, &isupipe.ReserveLivestreamRequest{
			Title:       name,
			Description: name,
			// FIXME: フロントで困らないようにちゃんとしたのを設定
			PlaylistUrl:  "",
			ThumbnailUrl: "",
			StartAt:      startAt.Unix(),
			EndAt:        endAt.Unix(),
			Tags:         []int64{},
		})
		if err != nil {
			overflow = true
		}
	}

	if !overflow {
		return fmt.Errorf("枠数を超過しても予約ができてしまいます")
	}

	return nil
}

func assertReserveOutOfTerm(ctx context.Context, testUser *isupipe.User, dnsResolver *resolver.DNSResolver) error {
	// 期間外の予約をするとエラーになる
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: testUser.Name,
		Password: "test",
	}); err != nil {
		return err
	}

	var (
		startAt = time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local)
		endAt   = time.Date(2026, 4, 1, 1, 0, 0, 0, time.Local)
	)
	if _, err := client.ReserveLivestream(ctx, testUser.Name, &isupipe.ReserveLivestreamRequest{
		Title:        "outofterm",
		Description:  "outofterm",
		PlaylistUrl:  "",
		ThumbnailUrl: "",
		StartAt:      startAt.Unix(),
		EndAt:        endAt.Unix(),
		Tags:         []int64{},
	}, isupipe.WithStatusCode(http.StatusBadRequest)); err != nil {
		return fmt.Errorf("期間外予約が不正にできてしまいます")
	}

	var (
		startAt2 = time.Date(2022, 4, 1, 0, 0, 0, 0, time.Local)
		endAt2   = time.Date(2022, 4, 1, 1, 0, 0, 0, time.Local)
	)
	if _, err := client.ReserveLivestream(ctx, testUser.Name, &isupipe.ReserveLivestreamRequest{
		Title:        "outofterm",
		Description:  "outofterm",
		PlaylistUrl:  "",
		ThumbnailUrl: "",
		StartAt:      startAt2.Unix(),
		EndAt:        endAt2.Unix(),
		Tags:         []int64{},
	}, isupipe.WithStatusCode(http.StatusBadRequest)); err != nil {
		return fmt.Errorf("期間外予約が不正にできてしまいます")
	}

	return nil
}

func assertMultipleEnterLivestream(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	return nil
}
