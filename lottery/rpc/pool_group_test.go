package rpc

import (
	"app/config"
	"app/poolgroup"
	"context"
	"encoding/json"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"lottery/dao"
	"micro_service/services"
	"sync"
	"testing"
	"time"
)

func TestFreeChainConsumesReserveAndRefundWithoutLosingPoolIdentity(t *testing.T) {
	previous := config.CfgIns
	config.CfgIns = &config.Configs{Lock: &sync.RWMutex{}}
	defer func() { config.CfgIns = previous }()
	old := &dao.PoolSession{Kind: "slots", Symbol: "a", PoolSymbol: "old-pool", Amount: "12.1234", Version: 3}
	changes, err := poolSessionChanges(&services.SlotsDoBetFreeReq{UserId: 1, GameId: 2, WinAmount: "4.0001", FlowControl: &services.SlotDoBetFlowControl{RefundPoolAmount: "1"}}, nil, old, 7, "a", nil)
	if err != nil || len(changes) != 1 || changes[0].Value == nil {
		t.Fatal(changes, err)
	}
	got := changes[0].Value
	if got.Amount != "7.1233" || got.PoolSymbol != "old-pool" || got.Version != 3 {
		t.Fatalf("lost reserve or original identity: %+v", got)
	}
	changes, err = poolSessionChanges(&services.SlotsDoBetFreeReq{UserId: 1, GameId: 2, WinAmount: "0", FlowControl: &services.SlotDoBetFlowControl{RefundPoolAmount: "12.1234"}}, nil, old, 7, "a", nil)
	if err != nil || changes[0].Value != nil {
		t.Fatal("fully refunded chain remains open")
	}
}
func TestSessionCompletionIsAtomicAndFenced(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer rdb.Close()
	ctx := context.Background()
	lease, err := poolgroup.Acquire(ctx, rdb, 7)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	keys := []string{lease.Key, poolgroup.Key(7, "sessions"), poolgroup.Key(7, "seen:test"), poolgroup.Key(7, "inflight")}
	_ = rdb.HSet(ctx, keys[3], "request", "data").Err()
	raw, _ := json.Marshal([][]string{{"session", `{"amount":"10"}`}})
	if err = finishPoolTracking.Run(ctx, rdb, keys, "wrong-owner", string(raw), "1", "request").Err(); err == nil {
		t.Fatal("stale owner changed sessions")
	}
	if n, _ := rdb.HLen(ctx, keys[3]).Result(); n != 1 {
		t.Fatal("stale owner cleared inflight marker")
	}
	if err = finishPoolTracking.Run(ctx, rdb, keys, lease.Token, string(raw), "1", "request").Err(); err != nil {
		t.Fatal(err)
	}
	if n, _ := rdb.HLen(ctx, keys[3]).Result(); n != 0 {
		t.Fatal("completed request remains inflight")
	}
	if n, _ := rdb.Exists(ctx, keys[2]).Result(); n != 1 {
		t.Fatal("completion tracking marker absent")
	}
}
func TestPoolHandlerPanicReturnsErrorForReconciliation(t *testing.T) {
	_, err := callPoolHandler(context.Background(), nil, func(context.Context, interface{}) (interface{}, error) { panic("pool read failed") })
	if status.Code(err) != codes.Internal {
		t.Fatal("panic escaped financial request wrapper", err)
	}
}

func TestPoolTrackingUsesOriginalIdempotencyIdentity(t *testing.T) {
	first := &services.SlotsDoBetReq{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: "r", FlowControl: &services.SlotDoBetFlowControl{RefundPoolAmount: "1"}}
	second := &services.SlotsDoBetReq{UserId: 1, AgentId: 7, GameId: 2, CurrencyType: "CNY", RoundId: "r", FlowControl: &services.SlotDoBetFlowControl{RefundPoolAmount: "2"}}
	if primaryPoolIdempotencyKey(first) != primaryPoolIdempotencyKey(second) {
		t.Fatal("attached flow control changes must not bypass the original request identity")
	}
	if primaryPoolIdempotencyKey(first) != buildIdempotencyKey("slotsDoBet", "1", "2", "CNY", "r") {
		t.Fatal("tracker identity differs from settlement identity")
	}
}
func TestPoolIdempotencyExpiresWithOriginalSettlementCache(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer rdb.Close()
	ctx := context.Background()
	if err := rdb.Set(ctx, "idem:test", `{"s":"signature","d":true,"p":"response"}`, time.Hour).Err(); err != nil {
		t.Fatal(err)
	}
	if done, err := poolIdempotencyDone(ctx, rdb, "idem:test"); err != nil || !done {
		t.Fatal(done, err)
	}
	server.FastForward(time.Hour + time.Second)
	if done, err := poolIdempotencyDone(ctx, rdb, "idem:test"); err != nil || done {
		t.Fatal("expired settlement cache still treated as a replay", done, err)
	}
}
