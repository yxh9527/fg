package rpc

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"micro_service/services"
	"reflect"
	"testing"
	"time"
)

func TestPoolRequestPlayerIDs(t *testing.T) {
	for _, tt := range []struct {
		name string
		req  interface{}
		want []uint32
	}{
		{"single", &services.SlotsDoBetReq{UserId: 7}, []uint32{7}},
		{"free", &services.SlotsDoBetFreeReq{UserId: 8}, []uint32{8}},
		{"fruit", &services.FruitDoBetReq{UserId: 9}, []uint32{9}},
		{"ug5", &services.Ug5GuessReq{UserId: 10}, []uint32{10}},
		{"luck5", &services.Luck5DealReq{UserId: 11}, []uint32{11}},
		{"batch", &services.FruitSettleRoundReq{Players: []*services.FruitSettlePlayer{{UserId: 1}, {UserId: 2}}}, []uint32{1, 2}},
		{"invalid batch", &services.FruitSettleRoundReq{Players: []*services.FruitSettlePlayer{nil}}, nil},
		{"invalid identity", &services.SlotsDoBetReq{}, nil},
		{"no player", &services.PoolAmountResultReq{}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := poolRequestPlayerIDs(tt.req); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("players = %v, want %v", got, tt.want)
			}
		})
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
