package config

import (
	"github.com/shopspring/decimal"
	"testing"
)

func TestGroupRoutingIsAgentScopedAndKeepsGameIdentity(t *testing.T) {
	cfg := newTestCfg()
	a := &Pool{Symbol: "a", GameId: 1, Name: "A", Pool: map[int32]*PoolItem{1: {Base: decimal.NewFromInt(10)}}}
	b := &Pool{Symbol: "b", GameId: 2, Name: "B", Pool: map[int32]*PoolItem{1: {Base: decimal.NewFromInt(20)}}}
	cfg.SetDefaultPool("a", a)
	cfg.SetDefaultPool("b", b)
	group := &PoolGroupSnapshot{ID: 3, AgentID: 7, Name: "G", Symbols: []string{"a", "b"}, Config: a, Version: 1}
	cfg.ReplacePoolGroups(7, []*PoolGroupSnapshot{group})
	if cfg.PoolStorageSymbol(7, "a") != cfg.PoolStorageSymbol(7, "b") {
		t.Fatal("group members have different storage")
	}
	if cfg.PoolStorageSymbol(8, "a") != "a" {
		t.Fatal("group leaked to another agent")
	}
	got := cfg.GetPoolCfg(7, "b")
	if got.Symbol != "b" || got.GameId != 2 || !got.Pool[1].Base.Equal(decimal.NewFromInt(10)) {
		t.Fatal("effective configuration lost game identity or group parameters")
	}
	if cfg.GetIndependentPoolCfg(7, "b") != b {
		t.Fatal("original configuration changed")
	}
	cfg.ReplacePoolGroups(7, nil)
	if cfg.GetPoolCfg(7, "b") != b {
		t.Fatal("leaving group did not restore independent configuration")
	}
}
func TestValidatePoolGroupLimitsAndDuplicateSymbols(t *testing.T) {
	for _, symbols := range [][]string{nil, {"a", "a"}, {""}, {"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}} {
		if ValidatePoolGroup("组", symbols) == nil {
			t.Fatalf("accepted invalid members: %v", symbols)
		}
	}
	if err := ValidatePoolGroup("组", []string{"a"}); err != nil {
		t.Fatal(err)
	}
	if ValidatePoolGroup(" ", []string{"a"}) == nil {
		t.Fatal("accepted empty name")
	}
}

func TestPoolLookupByGameIDUsesCurrentAgentAndGroup(t *testing.T) {
	cfg := newTestCfg()
	base := &Pool{Symbol: "a", GameId: 1, Pool: map[int32]*PoolItem{1: {Base: decimal.NewFromInt(10)}}}
	other := &Pool{Symbol: "a", GameId: 1, Pool: map[int32]*PoolItem{1: {Base: decimal.NewFromInt(999)}}}
	cfg.SetDefaultPool("a", base)
	cfg.SetAgentPool("8", other)
	if got := cfg.GetPoolCfgByGameId(7, 1); got != base {
		t.Fatal("looked up another agent's configuration")
	}
	cfg.ReplacePoolGroups(7, []*PoolGroupSnapshot{{ID: 3, AgentID: 7, Name: "G", Symbols: []string{"a"}, Config: other, Version: 1}})
	got := cfg.GetPoolCfgByGameId(7, 1)
	if got.Symbol != "a" || got.GameId != 1 || !got.Pool[1].Base.Equal(decimal.NewFromInt(999)) {
		t.Fatal("game ID lookup ignored group parameters")
	}
}
