package dao

import (
	"app/config"
	"app/poolgroup"
	"context"
	"encoding/json"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"micro_service/services"
	"sync"
	"testing"
)

func poolTestRuntime(t *testing.T) (context.Context, *poolgroup.Lease) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	oldRedis, oldCache, oldConfig := redisDao, gameCache, config.CfgIns
	redisDao = &RedisDao{cli: client}
	gameCache = &GameCacheMgr{lock: &sync.RWMutex{}, agents: map[int64]*AgentData{}}
	cfg := &config.Configs{Lock: &sync.RWMutex{}, Pool: &config.PoolMgr{Default: map[string]*config.Pool{}, Agent: map[string]*config.Pool{}}}
	config.CfgIns = cfg
	for i, symbol := range []string{"a", "b"} {
		cfg.SetDefaultPool(symbol, &config.Pool{Symbol: symbol, GameId: int64(i + 1), Pool: map[int32]*config.PoolItem{1: {Base: decimal.NewFromInt(100), Revenue: decimal.RequireFromString("0.1")}}})
	}
	group := &config.PoolGroupSnapshot{ID: 9, AgentID: 7, Name: "G", Symbols: []string{"a", "b"}, SourceSymbol: "a", Config: cfg.GetIndependentPoolCfg(7, "a"), Version: 1}
	state, _ := json.Marshal(&poolgroup.State{Groups: []*config.PoolGroupSnapshot{group}})
	ctx := context.Background()
	if err := client.Set(ctx, poolgroup.Key(7, "state"), state, 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.HSet(ctx, "player_1", "account", "tester", "isTourist", 0).Err(); err != nil {
		t.Fatal(err)
	}
	lease, err := poolgroup.Acquire(ctx, client, 7)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		lease.Close()
		_ = client.Close()
		redisDao = oldRedis
		gameCache = oldCache
		config.CfgIns = oldConfig
	})
	return ctx, lease
}
func TestGroupPoolSharedThroughRealCacheAndSurvivesReload(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	CacheIns().ChangePool(7, 1, "a", "CNY", "r1", decimal.NewFromInt(10), decimal.NewFromInt(2), decimal.RequireFromString("0.1"))
	if got := CacheIns().GetPool(7, "b"); !got.Equal(decimal.NewFromInt(107)) {
		t.Fatalf("other member sees %s, want 107", got)
	}
	if !CacheIns().CheckPoolWithChange(7, "b", "r2", "CNY", decimal.NewFromInt(3), decimal.Zero, decimal.Zero, 1) {
		t.Fatal("second member cannot use group pool")
	}
	if err := FlushPoolRequest(ctx, 7, lease); err != nil {
		t.Fatal(err)
	}
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if got := CacheIns().GetPool(7, "a"); !got.Equal(decimal.NewFromInt(104)) {
		t.Fatalf("persisted balance %s, want 104", got)
	}
	for _, symbol := range []string{"a", "b"} {
		if _, err := RedisIns().cli.ZScore(ctx, poolgroup.StatKeys()[0], poolgroup.Member(7, symbol)).Result(); err != redis.Nil {
			t.Fatalf("independent pool was recreated: %s %v", symbol, err)
		}
	}
	user := CacheIns().GetUser(7, 1)
	if !user.TotalEffectBet.Equal(decimal.NewFromInt(10)) {
		t.Fatal("player statistics did not survive cache refresh")
	}
}
func TestStaleTimerSnapshotCannotRecreateOriginalPool(t *testing.T) {
	ctx, _ := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	SaveAgentData([]*Game{{AgentId: 7, Symbol: "a", TotalEffectBet: decimal.NewFromInt(999), UpdateTime: 1}})
	if _, err := RedisIns().cli.ZScore(ctx, poolgroup.StatKeys()[0], poolgroup.Member(7, "a")).Result(); err != redis.Nil {
		t.Fatalf("stale snapshot recreated old pool: %v", err)
	}
}
func TestPoolFlushRejectsLostLease(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	CacheIns().ChangePool(7, 1, "a", "CNY", "r", decimal.NewFromInt(10), decimal.Zero, decimal.Zero)
	if err := RedisIns().cli.Set(ctx, lease.Key, "new-owner", 0).Err(); err != nil {
		t.Fatal(err)
	}
	if err := FlushPoolRequest(ctx, 7, lease); err == nil {
		t.Fatal("stale owner wrote group totals")
	}
	got, _ := RedisIns().cli.ZScore(ctx, poolgroup.StatKeys()[0], poolgroup.Member(7, config.PoolGroupSymbol(9))).Result()
	if got != 0 {
		t.Fatal("stale owner changed authoritative pool")
	}
}

func TestLuck5UsesOriginalGroupPoolAndDoesNotPayTwice(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if err := RedisIns().cli.HSet(ctx, "player_1", "currency", 10000).Err(); err != nil {
		t.Fatal(err)
	}
	deal := &Luck5Round{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: "luck5-group", Symbol: "b", BetCent: 100, ReserveCent: 200, Exchange: decimal.NewFromInt(1), RevenueRate: decimal.RequireFromString("0.1")}
	response, err := CacheIns().Luck5Deal(ctx, deal)
	if err != nil || response.Code != services.ErrorCode_OK {
		t.Fatal(response, err)
	}
	if deal.PoolSymbol != config.PoolGroupSymbol(9) {
		t.Fatal("deal did not capture original group identity")
	}
	if err = FlushPoolRequest(ctx, 7, lease); err != nil {
		t.Fatal(err)
	}
	if err = BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if response, err = CacheIns().Luck5Deal(ctx, deal); err != nil || response.CurrencyCent != 9900 {
		t.Fatal("deal replay changed wallet", response, err)
	}
	// The immutable round identity must survive even an unexpected mapping drift.
	config.CfgIns.ReplacePoolGroups(7, nil)
	settle := &Luck5Round{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: deal.RoundId, WinCent: 50, AttemptId: "attempt", Signature: "signature", RecordJson: "{}"}
	saveRecord := func(*Luck5Round) error { return nil }
	if response, err = CacheIns().Luck5Settle(ctx, settle, saveRecord); err != nil || !response.Ret || response.CurrencyCent != 9950 {
		t.Fatal("settle failed", response, err)
	}
	if response, err = CacheIns().Luck5Settle(ctx, settle, saveRecord); err != nil || response.CurrencyCent != 9950 {
		t.Fatal("settle replay paid twice", response, err)
	}
	if err = FlushPoolRequest(ctx, 7, lease); err != nil {
		t.Fatal(err)
	}
	if err = BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if got := CacheIns().GetPool(7, "a"); !got.Equal(decimal.RequireFromString("100.4")) {
		t.Fatal("settlement did not return to original group", got)
	}
	for _, symbol := range []string{"a", "b"} {
		if _, err := RedisIns().cli.ZScore(ctx, poolgroup.StatKeys()[0], poolgroup.Member(7, symbol)).Result(); err != redis.Nil {
			t.Fatal("Luck5 recreated original game pool", symbol, err)
		}
	}
}
