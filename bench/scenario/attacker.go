package scenario

import (
	"context"
	"sync"

	"github.com/isucon/isucon13/bench/internal/attacker"
)

func DnsWaterTortureAttackScenario(ctx context.Context, parallelism int) error {
	var attackGrp sync.WaitGroup
	for i := 0; i < parallelism; i++ {
		attackGrp.Add(1)
		go func() {
			defer attackGrp.Done()
			atk := attacker.NewDnsWaterTortureAttacker()
			atk.Attack(ctx)
		}()
	}
	attackGrp.Wait()

	return nil
}
