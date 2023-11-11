package scenario

import (
	"context"
	_ "embed"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/isupipe"
	"golang.org/x/sync/errgroup"
)

const testUserRawPassword = "s3cr3t"

func setupTestUser(ctx context.Context) (*isupipe.User, error) {
	client, err := isupipe.NewClient(agent.WithTimeout(config.PretestTimeout))
	if err != nil {
		return nil, err
	}

	user, err := client.Register(ctx, &isupipe.RegisterRequest{
		Name:     "pretestuser",
		Password: "test",
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

// 初期データチェック -> 基本的なエンドポイントの機能テスト -> 前後比較テスト
func Pretest(ctx context.Context) error {
	// 初期データチェック
	initialGrp, initialCtx := errgroup.WithContext(ctx)
	initialGrp.Go(func() error {
		return normalInitialPaymentPretest(initialCtx)
	})
	// initialGrp.Go(func() error {
	// 	return normalInitialLivecommentPretest(initialCtx)
	// })
	// initialGrp.Go(func() error {
	// 	return normalInitialReactionPretest(initialCtx)
	// })
	initialGrp.Go(func() error {
		return normalInitialTagPretest(ctx)
	})
	if err := initialGrp.Wait(); err != nil {
		return err
	}

	// 独立動作可能なテスト
	testUser, err := setupTestUser(ctx)
	if err != nil {
		return err
	}
	if err := NormalLivestreamPretest(ctx, testUser); err != nil {
		return err
	}

	logicGrp, childCtx := errgroup.WithContext(ctx)
	// 正常系
	logicGrp.Go(func() error {
		return NormalUserPretest(childCtx)
	})
	logicGrp.Go(func() error {
		return NormalIconPretest(childCtx)
	})
	logicGrp.Go(func() error {
		return NormalReactionPretest(childCtx, testUser)
	})
	logicGrp.Go(func() error {
		if err := NormalPostLivecommentPretest(childCtx, testUser); err != nil {
			return err
		}
		if err := NormalModerateLivecommentPretest(childCtx, testUser); err != nil {
			return err
		}
		return nil
	})
	// 異常系
	logicGrp.Go(func() error {
		return assertBadLogin(childCtx)
	})
	logicGrp.Go(func() error {
		return assertPipeUserRegistration(childCtx)
	})
	logicGrp.Go(func() error {
		return assertUserUniqueConstraint(childCtx)
	})
	logicGrp.Go(func() error {
		return assertReserveOverflowPretest(childCtx)
	})
	logicGrp.Go(func() error {
		return assertReserveOutOfTerm(childCtx, testUser)
	})
	logicGrp.Go(func() error {
		return assertMultipleEnterLivestream(childCtx)
	})
	if err := logicGrp.Wait(); err != nil {
		return err
	}

	// FIXME: 統計情報、Paymentなど、前後比較するものは他のシナリオが割り込まないようにする

	return nil
}
