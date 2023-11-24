package scenario

import (
	"context"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/isupipe"
	"go.uber.org/zap"
)

const testUserRawPassword = "s3cr3t"

func setupTestUser(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) (*isupipe.User, error) {
	client, err := isupipe.NewCustomResolverClient(
		contestantLogger,
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
func Pretest(ctx context.Context, contestantLogger *zap.Logger, dnsResolver *resolver.DNSResolver) error {
	// dns 初期レコード
	if err := dnsRecordPretest(ctx, dnsResolver); err != nil {
		return err
	}

	// 初期データチェック
	// FIXME: reactions, livecommentsは統計情報をもとにチェックする
	// FIXME: ngwordsはライブ配信のIDをいくつか問い合わせ、存在することをチェックする
	if err := normalInitialPaymentPretest(ctx, contestantLogger, dnsResolver); err != nil {
		return err
	}

	testUser, err := setupTestUser(ctx, contestantLogger, dnsResolver)
	if err != nil {
		return err
	}
	if err := NormalLivestreamPretest(ctx, contestantLogger, testUser, dnsResolver); err != nil {
		return err
	}

	// 正常系
	if err := NormalUserPretest(ctx, contestantLogger, dnsResolver); err != nil {
		return err
	}
	if err := NormalIconPretest(ctx, contestantLogger, dnsResolver); err != nil {
		return err
	}
	if err := NormalReactionPretest(ctx, contestantLogger, testUser, dnsResolver); err != nil {
		return err
	}
	if err := NormalPostLivecommentPretest(ctx, contestantLogger, testUser, dnsResolver); err != nil {
		return err
	}
	if err := NormalModerateLivecommentPretest(ctx, contestantLogger, testUser, dnsResolver); err != nil {
		return err
	}

	// 異常系
	if err := assertBadLogin(ctx, contestantLogger, dnsResolver); err != nil {
		return err
	}
	if err := assertPipeUserRegistration(ctx, contestantLogger, dnsResolver); err != nil {
		return err
	}
	if err := assertUserUniqueConstraint(ctx, contestantLogger, dnsResolver); err != nil {
		return err
	}
	if err := assertReserveOverflowPretest(ctx, contestantLogger, dnsResolver); err != nil {
		return err
	}
	if err := assertReserveOutOfTerm(ctx, contestantLogger, testUser, dnsResolver); err != nil {
		return err
	}
	if err := assertMultipleEnterLivestream(ctx, dnsResolver); err != nil {
		return err
	}

	return nil
}
