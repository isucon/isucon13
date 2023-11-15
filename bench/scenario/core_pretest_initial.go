package scenario

import (
	"context"
	"fmt"

	"github.com/isucon/isucandar/agent"
	"github.com/isucon/isucon13/bench/internal/config"
	"github.com/isucon/isucon13/bench/internal/resolver"
	"github.com/isucon/isucon13/bench/internal/scheduler"
	"github.com/isucon/isucon13/bench/isupipe"
)

// 初期データpretest

func normalInitialPaymentPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	// 初期状態で0円であるか
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	result, err := client.GetPaymentResult(ctx)
	if err != nil {
		return err
	}

	if result.TotalTip != 0 {
		return fmt.Errorf("初期の売上は0ISUでなければなりません")
	}

	return nil
}

func normalInitialLivecommentPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: "test001",
		Password: "test",
	}); err != nil {
		return err
	}

	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}
	livestream := livestreams[0]

	ngwords, err := client.GetNgwords(ctx, livestream.ID)
	if err != nil {
		return err
	}

	if len(ngwords) != config.InitialNgWords {
		return fmt.Errorf("初期データのNGワード数が不正です")
	}

	return nil
}

func normalInitialReactionPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: "test001",
		Password: "test",
	}); err != nil {
		return err
	}

	livestreams, err := client.GetMyLivestreams(ctx)
	if err != nil {
		return err
	}
	livestream := livestreams[0]

	reactions, err := client.GetReactions(ctx, livestream.ID)
	if err != nil {
		return err
	}

	if len(reactions) != config.InitialReactionCount {
		return fmt.Errorf("初期データのリアクション数が不正です")
	}

	return nil
}

func normalInitialTagPretest(ctx context.Context, dnsResolver *resolver.DNSResolver) error {
	// 初期データが期待する件数あるか
	client, err := isupipe.NewCustomResolverClient(
		dnsResolver,
		agent.WithTimeout(config.PretestTimeout),
	)
	if err != nil {
		return err
	}

	if err := client.Login(ctx, &isupipe.LoginRequest{
		Username: "test001",
		Password: "test",
	}); err != nil {
		return err
	}

	tagResponse, err := client.GetTags(ctx)
	if err != nil {
		return err
	}
	if len(tagResponse.Tags) != scheduler.GetTagPoolLength() {
		return fmt.Errorf("初期データのタグが正常に登録されていません: want=%d, but got=%d", scheduler.GetTagPoolLength(), len(tagResponse.Tags))
	}

	return nil
}
