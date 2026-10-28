package poolgroup

import (
	"app/config"
	"app/tables/manager"
	"context"
	"encoding/json"
	"errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"testing"
	"time"
)

func testGroup(id, agent int64, symbols ...string) *config.PoolGroupSnapshot {
	return &config.PoolGroupSnapshot{ID: id, AgentID: agent, Name: "组合池", SourceSymbol: symbols[0], Symbols: symbols, Version: 1, Config: &config.Pool{Symbol: symbols[0], Pool: map[int32]*config.PoolItem{1: {Base: decimal.NewFromInt(100)}}}}
}
func testStore(t *testing.T) (Store, context.Context, *Lease) {
	t.Helper()
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	lease, err := Acquire(ctx, rdb, 7)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(lease.Close)
	return Store{Redis: rdb}, ctx, lease
}
func seed(t *testing.T, s Store, ctx context.Context, agent int64, symbol string, amount int64) {
	t.Helper()
	for _, key := range StatKeys() {
		if err := s.Redis.ZAdd(ctx, key, redis.Z{Score: float64(amount), Member: Member(agent, symbol)}).Err(); err != nil {
			t.Fatal(err)
		}
	}
}
func exists(t *testing.T, s Store, ctx context.Context, agent int64, symbol string, want bool) {
	t.Helper()
	for _, key := range StatKeys() {
		_, err := s.Redis.ZScore(ctx, key, Member(agent, symbol)).Result()
		if want && err != nil {
			t.Fatalf("lost %s / %s: %v", key, symbol, err)
		}
		if !want && err != redis.Nil {
			t.Fatalf("stale %s / %s: %v", key, symbol, err)
		}
	}
}

func TestApplyCreateClearsOriginalPoolsButNotOtherAgent(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	restore := zap.ReplaceGlobals(zap.New(core))
	t.Cleanup(restore)
	s, ctx, lease := testStore(t)
	seed(t, s, ctx, 7, "a", 30)
	seed(t, s, ctx, 7, "b", 40)
	seed(t, s, ctx, 8, "a", 50)
	group := testGroup(12, 7, "a", "b")
	op := manager.PoolGroupOperation{ID: 1, AgentID: 7, GroupID: 12, Kind: "create"}
	state, err := s.Apply(ctx, lease, &State{}, op, Plan{Target: group})
	if err != nil {
		t.Fatal(err)
	}
	exists(t, s, ctx, 7, "a", false)
	exists(t, s, ctx, 7, "b", false)
	exists(t, s, ctx, 8, "a", true)
	if len(state.Groups) != 1 || !state.Groups[0].Config.Pool[1].Base.Equal(decimal.NewFromInt(100)) {
		t.Fatal("new pool must start at its copied base")
	}
	matched := false
	for _, entry := range logs.All() {
		fields := entry.ContextMap()
		if fields["stage"] == "redis_applied" {
			matched = true
			if fields["deletedRecordCount"] != int64(8) {
				t.Fatal("log does not report actual Redis deletion count", fields)
			}
		}
	}
	if !matched {
		t.Fatal("missing Redis applied stage log")
	}
}
func TestApplyMembershipEditPreservesGroupCountersAndConfig(t *testing.T) {
	s, ctx, lease := testStore(t)
	old := testGroup(12, 7, "a", "b")
	next := *old
	next.Symbols = []string{"b", "c"}
	next.Version = 2
	symbol := config.PoolGroupSymbol(12)
	seed(t, s, ctx, 7, symbol, 88)
	seed(t, s, ctx, 7, "a", 90)
	seed(t, s, ctx, 7, "c", 70)
	state := &State{Groups: []*config.PoolGroupSnapshot{old}, LastOperationID: 1}
	op := manager.PoolGroupOperation{ID: 2, AgentID: 7, GroupID: 12, Kind: "update"}
	updated, err := s.Apply(ctx, lease, state, op, Plan{Intent: Intent{Version: 1}, Target: &next})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range StatKeys() {
		got, err := s.Redis.ZScore(ctx, key, Member(7, symbol)).Result()
		if err != nil || got != 88 {
			t.Fatalf("group balance changed: %v %v", got, err)
		}
	}
	exists(t, s, ctx, 7, "a", false)
	exists(t, s, ctx, 7, "c", false)
	if updated.Groups[0].SourceSymbol != "a" || updated.Groups[0].Config != old.Config {
		t.Fatal("removing first member must not replace the copied configuration")
	}
}
func TestApplyRetryDoesNotClearSubsequentActivity(t *testing.T) {
	s, ctx, lease := testStore(t)
	group := testGroup(12, 7, "a")
	op := manager.PoolGroupOperation{ID: 1, AgentID: 7, GroupID: 12, Kind: "create"}
	plan := Plan{Target: group}
	state, err := s.Apply(ctx, lease, &State{}, op, plan)
	if err != nil {
		t.Fatal(err)
	}
	seed(t, s, ctx, 7, config.PoolGroupSymbol(12), 123)
	// Simulate retry with a stale caller snapshot after Redis applied, before SQL confirmation.
	if _, err = s.Apply(ctx, lease, &State{}, op, plan); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Apply(ctx, lease, state, op, plan); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Redis.ZScore(ctx, StatKeys()[0], Member(7, config.PoolGroupSymbol(12))).Result()
	if got != 123 {
		t.Fatal("retry reset an already active pool")
	}
}
func TestDeleteClearsGroupAndStartsMembersFresh(t *testing.T) {
	s, ctx, lease := testStore(t)
	group := testGroup(12, 7, "a", "b")
	seed(t, s, ctx, 7, config.PoolGroupSymbol(12), 123)
	seed(t, s, ctx, 7, "a", 99)
	updated, err := s.Apply(ctx, lease, &State{Groups: []*config.PoolGroupSnapshot{group}}, manager.PoolGroupOperation{ID: 2, AgentID: 7, GroupID: 12, Kind: "delete"}, Plan{Intent: Intent{Version: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Groups) != 0 {
		t.Fatal("deleted group remains routed")
	}
	exists(t, s, ctx, 7, config.PoolGroupSymbol(12), false)
	exists(t, s, ctx, 7, "a", false)
}
func TestResetKeepsDefinitionsAndClearsOnlyTargetAgent(t *testing.T) {
	s, ctx, lease := testStore(t)
	group := testGroup(12, 7, "a")
	seed(t, s, ctx, 7, config.PoolGroupSymbol(12), 44)
	seed(t, s, ctx, 7, "b", 55)
	seed(t, s, ctx, 8, "b", 66)
	updated, err := s.Apply(ctx, lease, &State{Groups: []*config.PoolGroupSnapshot{group}}, manager.PoolGroupOperation{ID: 3, AgentID: 7, Kind: "reset"}, Plan{Intent: Intent{Symbols: []string{"a", "b"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Groups) != 1 {
		t.Fatal("reset must preserve membership")
	}
	exists(t, s, ctx, 7, config.PoolGroupSymbol(12), false)
	exists(t, s, ctx, 7, "b", false)
	exists(t, s, ctx, 8, "b", true)
}
func TestLeaseLossRejectsClearWithoutPartialMutation(t *testing.T) {
	s, ctx, lease := testStore(t)
	seed(t, s, ctx, 7, "a", 10)
	_ = s.Redis.Set(ctx, lease.Key, "another-owner", 0).Err()
	_, err := s.Apply(ctx, lease, &State{}, manager.PoolGroupOperation{ID: 1, AgentID: 7, GroupID: 12, Kind: "create"}, Plan{Target: testGroup(12, 7, "a")})
	if err == nil {
		t.Fatal("lost lease allowed clearing pool")
	}
	exists(t, s, ctx, 7, "a", true)
}
func TestInvalidCounterTypeDoesNotPartiallyClear(t *testing.T) {
	s, ctx, lease := testStore(t)
	seed(t, s, ctx, 7, "a", 10)
	_ = s.Redis.Del(ctx, StatKeys()[3]).Err()
	_ = s.Redis.Set(ctx, StatKeys()[3], "corrupt", 0).Err()
	_, err := s.Apply(ctx, lease, &State{}, manager.PoolGroupOperation{ID: 1, AgentID: 7, GroupID: 12, Kind: "create"}, Plan{Target: testGroup(12, 7, "a")})
	if err == nil {
		t.Fatal("bad counter type accepted")
	}
	got, _ := s.Redis.ZScore(ctx, StatKeys()[0], Member(7, "a")).Result()
	if got != 10 {
		t.Fatal("partially cleared counters before validating types")
	}
}
func TestOperationGateCoversOldAndNewMembers(t *testing.T) {
	old := testGroup(12, 7, "a", "b")
	target := testGroup(12, 7, "b", "c")
	gate := OperationGate(&State{Groups: []*config.PoolGroupSnapshot{old}}, manager.PoolGroupOperation{ID: 5, GroupID: 12, Kind: "update"}, Plan{Target: target})
	for _, symbol := range []string{"a", "b", "c"} {
		if !gate.Affects(symbol) {
			t.Fatal("missing gate for", symbol)
		}
	}
	if gate.Affects("unrelated") {
		t.Fatal("unrelated game paused")
	}
	raw, _ := json.Marshal(gate)
	var decoded Gate
	if json.Unmarshal(raw, &decoded) != nil || decoded.OperationID != 5 {
		t.Fatal("gate serialization failed")
	}
}

func TestVirtualZeroSubmitUsesSharedLeaseBeforeSQL(t *testing.T) {
	store, ctx, _ := testStore(t)
	lease, err := Acquire(ctx, store.Redis, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	deadline, cancel := context.WithTimeout(ctx, 30*time.Millisecond)
	defer cancel()
	_, err = store.Submit(deadline, Intent{AgentID: 0, Kind: "create", RequestID: "zero-submit", Name: "zero", Symbols: []string{"a"}}, 1)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("test agent zero was rejected or bypassed the shared lease", err)
	}
}
func TestZeroAgentGroupCreateEditDeleteKeepsOtherAgentsIsolated(t *testing.T) {
	store, ctx, _ := testStore(t)
	lease, err := Acquire(ctx, store.Redis, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	seed(t, store, ctx, 0, "a", 10)
	seed(t, store, ctx, 0, "b", 20)
	seed(t, store, ctx, 7, "a", 90)
	target := testGroup(12, 0, "a")
	state, err := store.Apply(ctx, lease, &State{}, manager.PoolGroupOperation{ID: 1, AgentID: 0, GroupID: 12, Kind: "create"}, Plan{Target: target})
	if err != nil {
		t.Fatal(err)
	}
	exists(t, store, ctx, 0, "a", false)
	exists(t, store, ctx, 7, "a", true)
	seed(t, store, ctx, 0, config.PoolGroupSymbol(12), 75)
	updated := *target
	updated.Symbols = []string{"a", "b"}
	updated.Version = 2
	state, err = store.Apply(ctx, lease, state, manager.PoolGroupOperation{ID: 2, AgentID: 0, GroupID: 12, Kind: "update"}, Plan{Intent: Intent{Version: 1}, Target: &updated})
	if err != nil {
		t.Fatal(err)
	}
	exists(t, store, ctx, 0, "b", false)
	score, _ := store.Redis.ZScore(ctx, StatKeys()[0], Member(0, config.PoolGroupSymbol(12))).Result()
	if score != 75 {
		t.Fatal("editing zero agent group reset its balance", score)
	}
	state, err = store.Apply(ctx, lease, state, manager.PoolGroupOperation{ID: 3, AgentID: 0, GroupID: 12, Kind: "delete"}, Plan{Intent: Intent{Version: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Groups) != 0 {
		t.Fatal("zero agent group not deleted")
	}
	exists(t, store, ctx, 0, config.PoolGroupSymbol(12), false)
	exists(t, store, ctx, 7, "a", true)
}
