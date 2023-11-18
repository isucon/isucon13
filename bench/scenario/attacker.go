package scenario

import (
	"context"
	"net/http"

	"github.com/isucon/isucon13/bench/internal/attacker"
	"golang.org/x/time/rate"
)

var maxAttackOnEachScenario = 100

func DnsWaterTortureAttackScenario(ctx context.Context, httpClient *http.Client, loadLimiter *rate.Limiter) error {

	atk := attacker.NewDnsWaterTortureAttacker()
	for j := 0; j < maxAttackOnEachScenario; j++ {
		if err := loadLimiter.Wait(ctx); err == nil {
			atk.Attack(ctx, httpClient)
		}
	}

	return nil
}
