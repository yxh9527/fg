package dao

import (
	"fmt"
	"sync"
	"testing"

	"github.com/shopspring/decimal"
)

func TestTouristPoolAccrualExcluded(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := RedisIns().cli.HSet(ctx, "player_1", "isTourist", 1).Err(); err != nil {
		t.Fatal(err)
	}
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	before := CacheIns().GetPool(7, "a")
	if !CacheIns().ChangePool(7, 1, "a", "CNY", "tourist-round", decimal.NewFromInt(20), decimal.NewFromInt(5), decimal.RequireFromString("0.1")) {
		t.Fatal("tourist bet rejected")
	}
	CacheIns().Complete(7, 1, "a", decimal.NewFromInt(20), decimal.NewFromInt(5), decimal.RequireFromString("0.1"))
	if err := FlushPoolRequest(ctx, 7, lease); err != nil {
		t.Fatal(err)
	}
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if after := CacheIns().GetPool(7, "a"); !after.Equal(before) {
		t.Fatalf("tourist turnover changed persisted pool: before=%s after=%s", before, after)
	}
}

func TestTouristRefundExcluded(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := RedisIns().cli.HSet(ctx, "player_1", "isTourist", 1).Err(); err != nil {
		t.Fatal(err)
	}
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	before := CacheIns().GetPool(7, "a")
	CacheIns().ReturnPool(7, 1, "a", decimal.NewFromInt(10))
	if err := FlushPoolRequest(ctx, 7, lease); err != nil {
		t.Fatal(err)
	}
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if after := CacheIns().GetPool(7, "a"); !after.Equal(before) {
		t.Fatalf("tourist refund changed persisted pool: before=%s after=%s", before, after)
	}
}

func TestLuck5TouristRoundKeepsFlagAndPool(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := RedisIns().cli.HSet(ctx, "player_1", "isTourist", 1, "currency", 10000).Err(); err != nil {
		t.Fatal(err)
	}
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	before := CacheIns().GetPool(7, "a")
	deal := &Luck5Round{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: "tourist-luck5", Symbol: "b", BetCent: 100, ReserveCent: 200, Exchange: decimal.NewFromInt(1), RevenueRate: decimal.RequireFromString("0.1")}
	response, err := CacheIns().Luck5Deal(ctx, deal)
	if err != nil || response == nil || !response.Ret || deal.IsTourist != 1 {
		t.Fatalf("tourist deal failed: response=%+v err=%v flag=%d", response, err, deal.IsTourist)
	}
	if err := FlushPoolRequest(ctx, 7, lease); err != nil {
		t.Fatal(err)
	}
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	saved := false
	settle := &Luck5Round{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: deal.RoundId, WinCent: 50, RecordJson: "{}"}
	response, err = CacheIns().Luck5Settle(ctx, settle, func(round *Luck5Round) error {
		saved = true
		if round.IsTourist != 1 {
			t.Fatalf("settlement lost tourist flag: %d", round.IsTourist)
		}
		return nil
	})
	if err != nil || response == nil || !response.Ret || !saved {
		t.Fatalf("tourist settlement failed: response=%+v err=%v saved=%v", response, err, saved)
	}
	if err := FlushPoolRequest(ctx, 7, lease); err != nil {
		t.Fatal(err)
	}
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if after := CacheIns().GetPool(7, "a"); !after.Equal(before) {
		t.Fatalf("tourist Luck5 changed persisted pool: before=%s after=%s", before, after)
	}
}

func TestRealPlayerRefundStillReturnsPool(t *testing.T) {
	ctx, lease := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	CacheIns().ChangePool(7, 1, "a", "CNY", "real-round", decimal.NewFromInt(20), decimal.NewFromInt(15), decimal.Zero)
	before := CacheIns().GetPool(7, "a")
	CacheIns().ReturnPool(7, 1, "a", decimal.NewFromInt(10))
	if err := FlushPoolRequest(ctx, 7, lease); err != nil {
		t.Fatal(err)
	}
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if got := CacheIns().GetPool(7, "a"); !got.Equal(before.Add(decimal.NewFromInt(10))) {
		t.Fatalf("normal refund changed: before=%s after=%s", before, got)
	}
}

func TestInvalidPlayerIdentityIsNotCached(t *testing.T) {
	for _, managed := range []bool{false, true} {
		for _, invalid := range []string{"missing", "", "invalid", "2", "empty-account"} {
			t.Run(fmt.Sprintf("managed=%v/%s", managed, invalid), func(t *testing.T) {
				ctx, _ := poolTestRuntime(t)
				if err := RedisIns().cli.HSet(ctx, "player_1", "isTourist", invalid).Err(); err != nil {
					t.Fatal(err)
				}
				if invalid == "missing" {
					RedisIns().cli.HDel(ctx, "player_1", "isTourist")
				}
				if invalid == "empty-account" {
					RedisIns().cli.HSet(ctx, "player_1", "isTourist", 1, "account", "")
				}
				agent := &AgentData{Id: 7, poolManaged: managed, lock: &sync.RWMutex{}, userCache: make(map[uint32]*User)}
				failed := false
				func() {
					defer func() { failed = recover() != nil }()
					agent.GetUser(1)
				}()
				if !failed || len(agent.userCache) != 0 {
					t.Fatal("invalid player identity was accepted or cached")
				}
				if err := RedisIns().cli.HSet(ctx, "player_1", "account", "demo", "isTourist", 1).Err(); err != nil {
					t.Fatal(err)
				}
				if user := agent.GetUser(1); user.IsTourist != 1 || user.Account != "demo" {
					t.Fatalf("retry lost tourist identity: %+v", user)
				}
			})
		}
	}
}
