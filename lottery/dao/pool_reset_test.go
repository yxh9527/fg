package dao

import (
	"app/config"
	"app/poolgroup"
	"app/tables/manager"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"testing"
)

func TestResetReplayPreservesNewRedisAndMemoryActivity(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	state, err := PoolGroupStore().Load(ctx, 7)
	if err != nil {
		t.Fatal(err)
	}
	op := manager.PoolGroupOperation{ID: 10, AgentID: 7, Kind: "reset"}
	plan := poolgroup.Plan{}
	state, err = applyPoolRuntime(ctx, lease, state, op, plan)
	if err != nil {
		t.Fatal(err)
	}
	CacheIns().ChangePool(7, 1, "a", "CNY", "new-bet", decimal.NewFromInt(5), decimal.Zero, decimal.Zero)
	if err = FlushPoolRequest(ctx, 7, lease); err != nil {
		t.Fatal(err)
	}
	if _, err = applyPoolRuntime(ctx, lease, state, op, plan); err != nil {
		t.Fatal(err)
	}
	if got := CacheIns().GetPool(7, "b"); !got.Equal(decimal.NewFromInt(105)) {
		t.Fatal("duplicate reset erased new activity", got)
	}
	got, _ := RedisIns().cli.ZScore(ctx, poolgroup.StatKeys()[0], poolgroup.Member(7, config.PoolGroupSymbol(9))).Result()
	if got != 5 {
		t.Fatal("duplicate reset cleared Redis activity")
	}
}

func TestResetFailureDoesNotPretendMemoryWasCleared(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	agent := CacheIns().GetAgent(7)
	agent.lock.Lock()
	cached := agent.GetGame("a")
	cached.TotalEffectBet = decimal.NewFromInt(50)
	agent.lock.Unlock()
	if err := RedisIns().cli.ZAdd(ctx, poolgroup.StatKeys()[0], redis.Z{Score: 50, Member: poolgroup.Member(7, cached.Symbol)}).Err(); err != nil {
		t.Fatal(err)
	}
	if err := RedisIns().cli.Set(ctx, lease.Key, "other-owner", 0).Err(); err != nil {
		t.Fatal(err)
	}
	state, err := PoolGroupStore().Load(ctx, 7)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = applyPoolRuntime(ctx, lease, state, manager.PoolGroupOperation{ID: 10, AgentID: 7, Kind: "reset"}, poolgroup.Plan{}); err == nil {
		t.Fatal("lost lease reported reset success")
	}
	if !cached.TotalEffectBet.Equal(decimal.NewFromInt(50)) {
		t.Fatal("cleared memory before confirming Redis reset")
	}
}

func TestAlreadyAppliedResetInvalidatesOtherInstanceCache(t *testing.T) {
	ctx, _ := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	agent := CacheIns().GetAgent(7)
	agent.lock.Lock()
	cached := agent.GetGame("a")
	cached.TotalEffectBet = decimal.NewFromInt(50)
	cached.UpdateTime = 1
	agent.lock.Unlock()
	state, err := PoolGroupStore().Load(ctx, 7)
	if err != nil {
		t.Fatal(err)
	}
	state.Revision++
	state.LastOperationID = 10
	syncPoolRuntimeMemory(7, state)
	if len(agent.gameCache) != 0 || !cached.TotalEffectBet.IsZero() || cached.UpdateTime != 0 {
		t.Fatal("subscriber retained pre-reset in-memory data")
	}
}
