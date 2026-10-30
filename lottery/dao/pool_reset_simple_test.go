package dao

import (
	"app/config"
	"app/entity"
	"app/poolgroup"
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"testing"
)

func TestSimpleResetClearsOnlyPoolMemoryAndFourRedisKeys(t *testing.T) {
	ctx, _ := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	agent := CacheIns().GetAgent(7)
	cached := agent.GetGame("a")
	cached.TotalEffectBet = decimal.NewFromInt(50)
	cached.TotalChips = decimal.NewFromInt(60)
	cached.TotalProfLoss = decimal.NewFromInt(4)
	cached.TotalRevenue = decimal.NewFromInt(2)
	cached.UpdateTime = 1
	user := &User{UserId: 1, TotalEffectBet: decimal.NewFromInt(15)}
	agent.userCache[1] = user
	round := &RoundItem{AgentId: 7, RoundId: "unfinished"}
	agent.RoundCache[round.RoundId] = round
	other := CacheIns().GetAgent(8)
	otherCached := &Game{AgentId: 8, Symbol: "a", TotalEffectBet: decimal.NewFromInt(20), UpdateTime: 1}
	other.gameCache["a"] = otherCached
	rdb := RedisIns().cli
	for _, key := range poolgroup.StatKeys() {
		if err := rdb.ZAdd(ctx, key, redis.Z{Member: poolgroup.Member(7, "a"), Score: 50}, redis.Z{Member: poolgroup.Member(7, config.PoolGroupSymbol(9)), Score: 30}, redis.Z{Member: poolgroup.Member(8, "removed-game"), Score: 90}).Err(); err != nil {
			t.Fatal(err)
		}
	}
	if err := rdb.HSet(ctx, "player_1", "currency", 1234).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.ZAdd(ctx, "userTotalEffBet", redis.Z{Member: "1", Score: 15}).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.HSet(ctx, "rund_records", "unfinished", `{"agentId":7}`).Err(); err != nil {
		t.Fatal(err)
	}
	if err := rdb.HSet(ctx, "lottery:reservation_ledger", "example", "unchanged").Err(); err != nil {
		t.Fatal(err)
	}
	originalState, err := rdb.Get(ctx, poolgroup.Key(7, "state")).Result()
	if err != nil {
		t.Fatal(err)
	}
	if err = resetPoolCounters(ctx); err != nil {
		t.Fatal(err)
	}
	for _, key := range poolgroup.StatKeys() {
		if exists, err := rdb.Exists(ctx, key).Result(); err != nil || exists != 0 {
			t.Fatal("pool aggregate not deleted", key, exists, err)
		}
	}
	if !cached.TotalEffectBet.IsZero() || !cached.TotalChips.IsZero() || !cached.TotalProfLoss.IsZero() || !cached.TotalRevenue.IsZero() || cached.UpdateTime != 0 || len(agent.gameCache) != 0 {
		t.Fatal("old local snapshot can revive pool counters")
	}
	if !otherCached.TotalEffectBet.IsZero() || otherCached.UpdateTime != 0 || len(other.gameCache) != 0 {
		t.Fatal("reset missed another cached agent")
	}
	if agent.userCache[1] != user || len(agent.RoundCache) != 1 || round.Over {
		t.Fatal("pool reset modified player or round state")
	}
	if got, _ := rdb.Get(ctx, poolgroup.Key(7, "state")).Result(); got != originalState {
		t.Fatal("reset rewrote group definitions or added reset generation")
	}
	if got, _ := rdb.HGet(ctx, "player_1", "currency").Result(); got != "1234" {
		t.Fatal("wallet modified", got)
	}
	if got, _ := rdb.ZScore(ctx, "userTotalEffBet", "1").Result(); got != 15 {
		t.Fatal("player statistics modified", got)
	}
	if got, _ := rdb.HGet(ctx, "lottery:reservation_ledger", "example").Result(); got != "unchanged" {
		t.Fatal("unrelated key modified", got)
	}
	if exists, _ := rdb.HExists(ctx, "rund_records", "unfinished").Result(); !exists {
		t.Fatal("round data removed")
	}
	for _, part := range []string{"sessions", "inflight", "reset_rounds"} {
		if n, _ := rdb.Exists(ctx, poolgroup.Key(7, part)).Result(); n != 0 {
			t.Fatal("reset created tracking key", part)
		}
	}
	SaveAgentData([]*Game{cached, otherCached})
	for _, key := range poolgroup.StatKeys() {
		if n, _ := rdb.Exists(ctx, key).Result(); n != 0 {
			t.Fatal("old timer snapshot restored cleared pool", key)
		}
	}
	if got := CacheIns().GetPool(7, "a"); !got.Equal(decimal.NewFromInt(100)) {
		t.Fatal("pool no longer uses configured Base", got)
	}
}

func TestSimpleResetMessageAcceptsEmptyPayloadAndNoMysql(t *testing.T) {
	ctx, _ := poolTestRuntime(t)
	for _, payload := range []interface{}{nil, &entity.ResetPool{}} {
		if err := RedisIns().cli.ZAdd(ctx, poolgroup.StatKeys()[0], redis.Z{Member: "0_lucky", Score: 25}).Err(); err != nil {
			t.Fatal(err)
		}
		ResetPool(payload)
		if n, _ := RedisIns().cli.Exists(ctx, poolgroup.StatKeys()[0]).Result(); n != 0 {
			t.Fatal("empty legacy message did not reset pool")
		}
	}
}

func TestSimpleResetZerosMemoryBeforeRedisFailure(t *testing.T) {
	_, _ = poolTestRuntime(t)
	cached := &Game{AgentId: 7, Symbol: "a", TotalEffectBet: decimal.NewFromInt(50), UpdateTime: 1}
	CacheIns().GetAgent(7).gameCache["a"] = cached
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := resetPoolCounters(ctx); err == nil {
		t.Fatal("Redis failure was not reported")
	}
	if !cached.TotalEffectBet.IsZero() || cached.UpdateTime != 0 {
		t.Fatal("reset did not clear memory before deleting Redis")
	}
	if len(CacheIns().GetAgent(7).gameCache) != 0 {
		t.Fatal("cache lock was not released on failure")
	}
}

func TestSimpleResetIgnoresOldTrackingRecords(t *testing.T) {
	ctx, _ := poolTestRuntime(t)
	for _, part := range []string{"sessions", "inflight", "reset_rounds"} {
		if err := RedisIns().cli.Set(ctx, poolgroup.Key(7, part), "old-invalid-tracking-data", 0).Err(); err != nil {
			t.Fatal(err)
		}
	}
	if err := RedisIns().cli.ZAdd(ctx, poolgroup.StatKeys()[0], redis.Z{Member: "7_a", Score: 50}).Err(); err != nil {
		t.Fatal(err)
	}
	ResetPool(&entity.ResetPool{})
	if n, _ := RedisIns().cli.Exists(ctx, poolgroup.StatKeys()[0]).Result(); n != 0 {
		t.Fatal("old tracking data blocked reset")
	}
	for _, part := range []string{"sessions", "inflight", "reset_rounds"} {
		if got, _ := RedisIns().cli.Get(ctx, poolgroup.Key(7, part)).Result(); got != "old-invalid-tracking-data" {
			t.Fatal("reset changed data outside the four aggregates", part, got)
		}
	}
}
