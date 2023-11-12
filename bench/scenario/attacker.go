package scenario

import (
	"context"

	"github.com/isucon/isucon13/bench/internal/attacker"
	"golang.org/x/time/rate"
)

var maxAttackOnEachScenario = 1000

func DnsWaterTortureAttackScenario(ctx context.Context, loadLimiter *rate.Limiter) error {

	atk := attacker.NewDnsWaterTortureAttacker()
	for j := 0; j < maxAttackOnEachScenario; j++ {
		if err := loadLimiter.Wait(ctx); err == nil {
			atk.Attack(ctx)
		}
	}

	return nil
}
