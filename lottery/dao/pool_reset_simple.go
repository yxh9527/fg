package dao

import (
	"app/esindex"
	"context"

	"github.com/shopspring/decimal"
)

// Reset only pool counters: clear local snapshots first, then remove the four
// Redis aggregates. Configuration, wallets, player statistics and rounds remain.
func resetPoolCounters(ctx context.Context) error {
	if gameCache != nil {
		gameCache.lock.Lock()
		defer gameCache.lock.Unlock()
		locked := make([]*AgentData, 0, len(gameCache.agents))
		for _, agent := range gameCache.agents {
			agent.lock.Lock()
			locked = append(locked, agent)
		}
		defer func() {
			for i := len(locked) - 1; i >= 0; i-- {
				locked[i].lock.Unlock()
			}
		}()
		for _, agent := range locked {
			for _, game := range agent.gameCache {
				game.TotalEffectBet = decimal.Zero
				game.TotalChips = decimal.Zero
				game.TotalProfLoss = decimal.Zero
				game.TotalRevenue = decimal.Zero
				game.UpdateTime = 0
			}
			agent.gameCache = make(map[string]*Game)
		}
	}
	return RedisIns().cli.Del(ctx, esindex.AgentEffectData(), esindex.AgentChipsData(), esindex.AgentProfitLossData(), esindex.AgentRevenueData()).Err()
}
