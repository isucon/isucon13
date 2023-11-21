package scenario

import (
	"context"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/isupipe"
	"golang.org/x/sync/errgroup"
)

const testUserRawPassword = "s3cr3t"

func setupTestUser(ctx context.Context, dnsResolver *resolver.DNSResolver) (*isupipe.User, error) {
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return nil, err
	}

	user, err := client.Register(ctx, &isupipe.RegisterRequest{
		Name:        "pretestuser",
		Password:    "test",
		DisplayName: "pretest user",
		Description: "this is a pre test user",
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// 初期データチェック -> 基本的なエンドポイントの機能テスト -> 前後比較テスト
func Pretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	// dns 初期レコード
	if err := dnsRecordPretest(ctx, dnsResolver); err != nil {
		return err
	}

	// 初期データチェック
	initialGrp, initialCtx := errgroup.WithContext(ctx)
	initialGrp.Go(func() error {
		return normalInitialPaymentPretest(initialCtx, dnsResolver)
	})
	// FIXME: reactions, livecommentsは統計情報をもとにチェックする
	// FIXME: ngwordsはライブ配信のIDをいくつか問い合わせ、存在することをチェックする
	if err := initialGrp.Wait(); err != nil {
		return err
	}

	testUser, err := setupTestUser(ctx, dnsResolver)
	if err != nil {
		return err
	}
	if err := NormalLivestreamPretest(ctx, testUser, dnsResolver); err != nil {
		return err
	}

	// 正常系
	if err := NormalUserPretest(ctx, dnsResolver); err != nil {
		return err
	}
	if err := NormalIconPretest(ctx, dnsResolver); err != nil {
		return err
	}
	if err := NormalReactionPretest(ctx, testUser, dnsResolver); err != nil {
		return err
	}
	if err := NormalPostLivecommentPretest(ctx, testUser, dnsResolver); err != nil {
		return err
	}
	if err := NormalModerateLivecommentPretest(ctx, testUser, dnsResolver); err != nil {
		return err
	}

	// 異常系
	if err := assertBadLogin(ctx, dnsResolver); err != nil {
		return err
	}
	if err := assertPipeUserRegistration(ctx, dnsResolver); err != nil {
		return err
	}
	if err := assertUserUniqueConstraint(ctx, dnsResolver); err != nil {
		return err
	}
	if err := assertReserveOverflowPretest(ctx, dnsResolver); err != nil {
		return err
	}
	if err := assertReserveOutOfTerm(ctx, testUser, dnsResolver); err != nil {
		return err
	}
	if err := assertMultipleEnterLivestream(ctx, dnsResolver); err != nil {
		return err
	}

	return nil
}
