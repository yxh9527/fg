package dao

import (
	"app/config"
	"app/poolgroup"
	"app/tables/manager"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"testing"
)

type resetTestRepository struct {
	queued     string
	operations []manager.PoolGroupOperation
	queueError error
}

func (r *resetTestRepository) QueueReset(_ context.Context, batch string) error {
	r.queued = batch
	return r.queueError
}
func (r *resetTestRepository) ResetOperations(_ context.Context, _ string) ([]manager.PoolGroupOperation, error) {
	return r.operations, nil
}

func TestResetBatchActuallyClearsRedisAndMemoryBeforeReturning(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	agent := CacheIns().GetAgent(7)
	agent.lock.Lock()
	cached := agent.GetGame("a")
	cached.TotalEffectBet = decimal.NewFromInt(50)
	cached.TotalProfLoss = decimal.NewFromInt(4)
	cached.TotalRevenue = decimal.NewFromInt(2)
	cached.TotalChips = decimal.NewFromInt(60)
	cached.UpdateTime = 1
	agent.lock.Unlock()
	for _, key := range poolgroup.StatKeys() {
		for _, symbol := range []string{"a", "b", config.PoolGroupSymbol(9), "removed_game"} {
			if err := RedisIns().cli.ZAdd(ctx, key, redis.Z{Score: 50, Member: poolgroup.Member(7, symbol)}).Err(); err != nil {
				t.Fatal(err)
			}
		}
		if err := RedisIns().cli.ZAdd(ctx, key, redis.Z{Score: 90, Member: poolgroup.Member(8, "a")}).Err(); err != nil {
			t.Fatal(err)
		}
	}
	state, err := PoolGroupStore().Load(ctx, 7)
	if err != nil {
		t.Fatal(err)
	}
	repository := &resetTestRepository{operations: []manager.PoolGroupOperation{{ID: 10, AgentID: 7, Kind: "reset", Status: poolgroup.Pending}}}
	calls := 0
	err = executePoolResetBatch(ctx, repository, "reset-test", func(ctx context.Context, op manager.PoolGroupOperation) error {
		calls++
		var err error
		state, err = applyPoolRuntime(ctx, lease, state, op, poolgroup.Plan{Intent: poolgroup.Intent{Symbols: []string{"a"}}})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if repository.queued != "reset-test" || calls != 1 {
		t.Fatal("reset stopped after SQL enqueue without dispatching execution")
	}
	for _, key := range poolgroup.StatKeys() {
		members, err := RedisIns().cli.ZRange(ctx, key, 0, -1).Result()
		if err != nil || len(members) != 1 || members[0] != poolgroup.Member(8, "a") {
			t.Fatalf("Redis reset incomplete or cleared another agent: %s %v %v", key, members, err)
		}
	}
	if len(agent.gameCache) != 0 || !cached.TotalEffectBet.IsZero() || !cached.TotalProfLoss.IsZero() || !cached.TotalRevenue.IsZero() || !cached.TotalChips.IsZero() || cached.UpdateTime != 0 {
		t.Fatal("Redis was reset but original in-memory data remained")
	}
	if config.CfgIns.GetPoolGroup(7, "a") == nil {
		t.Fatal("reset removed group configuration")
	}
	if got := CacheIns().GetPool(7, "a"); !got.Equal(decimal.NewFromInt(100)) {
		t.Fatal("pool did not restart from Base", got)
	}
	SaveAgentData([]*Game{cached})
	for _, key := range poolgroup.StatKeys() {
		got, _ := RedisIns().cli.ZScore(ctx, key, poolgroup.Member(7, config.PoolGroupSymbol(9))).Result()
		if got != 0 {
			t.Fatal("old snapshot resurrected pre-reset data")
		}
	}
}

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

func TestResetWaitingIsReportedAndOtherAgentsAreAttempted(t *testing.T) {
	repository := &resetTestRepository{operations: []manager.PoolGroupOperation{{AgentID: 7}, {AgentID: 8}}}
	attempted := []int64{}
	err := executePoolResetBatch(context.Background(), repository, "batch", func(_ context.Context, op manager.PoolGroupOperation) error {
		attempted = append(attempted, op.AgentID)
		if op.AgentID == 7 {
			return errPoolResetWaiting
		}
		return nil
	})
	if !errors.Is(err, errPoolResetWaiting) || len(attempted) != 2 {
		t.Fatal("waiting reset was reported successful or blocked other agents", err, attempted)
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
