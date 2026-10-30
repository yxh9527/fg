package rpc

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"micro_service/services"
	"testing"
	"time"
)

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
