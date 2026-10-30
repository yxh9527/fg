package dao

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"micro_service/services"
	"testing"
	"time"
)

func TestLuck5KeepsOnlyOpenDealContextAndRemovesItAfterSettlement(t *testing.T) {
	ctx, _ := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if err := RedisIns().cli.HSet(ctx, "player_1", "currency", 10000).Err(); err != nil {
		t.Fatal(err)
	}
	deal := &Luck5Round{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: "normal-context", Symbol: "a", BetCent: 100, ReserveCent: 200, Exchange: decimal.NewFromInt(1)}
	response, err := CacheIns().Luck5Deal(ctx, deal)
	if err != nil || response.Code != services.ErrorCode_OK {
		t.Fatal(response, err)
	}
	raw, err := RedisIns().cli.Get(ctx, luck5Key(deal.RoundId)).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if ttl, err := RedisIns().cli.TTL(ctx, luck5Key(deal.RoundId)).Result(); err != nil || ttl != time.Hour {
		t.Fatalf("deal context TTL = %v, want %v; error: %v", ttl, time.Hour, err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Attempts", "FundsDone", "RecordDone", "Signature", "attempts", "fundsDone", "recordDone", "signature"} {
		if _, exists := fields[name]; exists {
			t.Fatal("dedup state was saved", name)
		}
	}
	round, err := RedisIns().loadLuck5Round(ctx, deal.RoundId)
	if err != nil || round.BetCent != 100 || round.ReserveCent != 200 {
		t.Fatal("settlement inputs missing", round, err)
	}
	saved := false
	response, err = CacheIns().Luck5Settle(ctx, &Luck5Round{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: deal.RoundId, WinCent: 50, RecordJson: "{}"}, func(round *Luck5Round) error {
		saved = true
		if round.BetCent != 100 || round.ReserveCent != 200 {
			t.Fatal(round)
		}
		return nil
	})
	if err != nil || response.Code != services.ErrorCode_OK || response.CurrencyCent != 9950 || !saved {
		t.Fatal(response, saved, err)
	}
	if n, _ := RedisIns().cli.Exists(ctx, luck5Key(deal.RoundId)).Result(); n != 0 {
		t.Fatal("finished deal context was retained")
	}
	for _, pattern := range []string{"fg_pool_group:*:seen:*", "fg_pool_group:*:sessions", "fg_pool_group:*:inflight"} {
		if keys, err := RedisIns().cli.Keys(ctx, pattern).Result(); err != nil || len(keys) != 0 {
			t.Fatal(pattern, keys, err)
		}
	}
}

func TestLuck5DoesNotCacheRepeatedDealOrRejectedCandidate(t *testing.T) {
	ctx, _ := poolTestRuntime(t)
	if err := BeginPoolRequest(ctx, 7); err != nil {
		t.Fatal(err)
	}
	if err := RedisIns().cli.HSet(ctx, "player_1", "currency", 10000).Err(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		request := &Luck5Round{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: "caller-controls-repeats", Symbol: "a", BetCent: 100, ReserveCent: 200, Exchange: decimal.NewFromInt(1)}
		response, err := CacheIns().Luck5Deal(ctx, request)
		if err != nil || response.CurrencyCent != int64(9900-100*i) {
			t.Fatal("lottery returned a cached deal result", response, err)
		}
	}
	rejected := &Luck5Round{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: "rejected-context", Symbol: "a", BetCent: 100, ReserveCent: 1000000, Exchange: decimal.NewFromInt(1)}
	response, err := CacheIns().Luck5Deal(ctx, rejected)
	if err != nil || response.Code != services.ErrorCode_NO_ENOUGH_POOL_MONEY {
		t.Fatal(response, err)
	}
	if n, _ := RedisIns().cli.Exists(ctx, luck5Key(rejected.RoundId)).Result(); n != 0 {
		t.Fatal("rejected candidate history was persisted")
	}
}
