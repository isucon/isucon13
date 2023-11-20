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
	if err := dnsrecordPretest(ctx, dnsResolver); err != nil {
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

	// 独立動作可能なテスト
	testUser, err := setupTestUser(ctx, dnsResolver)
	if err != nil {
		return err
	}
	if err := NormalLivestreamPretest(ctx, testUser, dnsResolver); err != nil {
		return err
	}

	logicGrp, childCtx := errgroup.WithContext(ctx)
	// 正常系
	logicGrp.Go(func() error {
		return NormalUserPretest(childCtx, dnsResolver)
	})
	logicGrp.Go(func() error {
		return NormalIconPretest(childCtx, dnsResolver)
	})
	logicGrp.Go(func() error {
		return NormalReactionPretest(childCtx, testUser, dnsResolver)
	})
	logicGrp.Go(func() error {
		if err := NormalPostLivecommentPretest(childCtx, testUser, dnsResolver); err != nil {
			return err
		}
		if err := NormalModerateLivecommentPretest(childCtx, testUser, dnsResolver); err != nil {
			return err
		}
		return nil
	})
	// 異常系
	logicGrp.Go(func() error {
		return assertBadLogin(childCtx, dnsResolver)
	})
	logicGrp.Go(func() error {
		return assertPipeUserRegistration(childCtx, dnsResolver)
	})
	logicGrp.Go(func() error {
		return assertUserUniqueConstraint(childCtx, dnsResolver)
	})
	logicGrp.Go(func() error {
		return assertReserveOverflowPretest(childCtx, dnsResolver)
	})
	logicGrp.Go(func() error {
		return assertReserveOutOfTerm(childCtx, testUser, dnsResolver)
	})
	logicGrp.Go(func() error {
		return assertMultipleEnterLivestream(childCtx, dnsResolver)
	})
	if err := logicGrp.Wait(); err != nil {
		return err
	}

	// FIXME: 統計情報、Paymentなど、前後比較するものは他のシナリオが割り込まないようにする

	return nil
}
