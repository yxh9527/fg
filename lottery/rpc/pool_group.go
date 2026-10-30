package rpc

import (
	"app/poolgroup"
	"app/poollog"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"lottery/dao"
	"micro_service/services"
	"time"
)

type poolRequestIdentity interface {
	GetAgentId() uint32
	GetGameId() uint32
}

// Pool requests only refresh group routing and persist existing counters.
// No auxiliary session, inflight, reset-round or seen records are created.
func poolGroupInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response interface{}, err error) {
	poolLogFields := append(poollog.MetadataFields(req), zap.String("rpcMethod", info.FullMethod))
	poolLogTrace := poollog.Begin("rpc.poolGroupInterceptor", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&err)
	identity, ok := req.(poolRequestIdentity)
	if !ok {
		return handler(ctx, req)
	}
	agentID := int64(identity.GetAgentId())
	acquireCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	lease, err := poolgroup.Acquire(acquireCtx, dao.PoolGroupStore().Redis, agentID)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "水池正忙，请重试")
	}
	defer lease.Close()
	if err = dao.BeginPoolRequest(ctx, agentID); err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	if game := dao.GamesManagerIns().GetById(int64(identity.GetGameId())); game != nil {
		if err = dao.PoolGroupStore().CheckInitialization(ctx, agentID, game.ConfName); err != nil {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
	}
	response, handlerErr := callPoolHandler(ctx, req, handler)
	finishCtx, finishCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer finishCancel()
	if err = dao.FlushPoolRequest(finishCtx, agentID, lease); err != nil {
		return nil, status.Error(codes.Unavailable, "水池累计保存失败："+err.Error())
	}
	return response, handlerErr
}

func primaryPoolIdempotencyKey(req interface{}) (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(req)...)
	poolLogTrace := poollog.BeginDebug("rpc.primaryPoolIdempotencyKey", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	switch r := req.(type) {
	case *services.SlotsDoBetReq:
		return buildIdempotencyKey("slotsDoBet", u32Str(r.UserId), u32Str(r.GameId), r.CurrencyType, r.RoundId)
	case *services.SlotsDoBetFreeReq:
		return buildIdempotencyKey("slotsDoBetFree", u32Str(r.UserId), u32Str(r.GameId), r.CurrencyType, r.RoundId)
	case *services.FruitDoBetReq:
		return buildIdempotencyKey("fruitDoBet", u32Str(r.UserId), u32Str(r.GameId), r.CurrencyType, r.RoundId)
	case *services.Ug5GuessReq:
		if r.Complete {
			return buildIdempotencyKey("ug5Guess", u32Str(r.UserId), u32Str(r.GameId), r.CurrencyType, r.RoundId)
		}
	case *services.FruitSettleRoundReq:
		return buildIdempotencyKey("fruitSettleRound", u32Str(r.AgentId), u32Str(r.GameId), r.CurrencyType, i64Str(r.Period), r.RoundId)
	}
	return ""
}
func poolIdempotencyDone(ctx context.Context, rdb redis.UniversalClient, key string) (poolLogResult0 bool, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.Begin("rpc.poolIdempotencyDone", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	raw, err := rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		poollog.Event(zap.ErrorLevel, "rpc.poolIdempotencyDone", "rdb.Get.Bytes", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return false, err
	}
	var entry idempotencyEntry
	if err = json.Unmarshal(raw, &entry); err != nil {
		poollog.Event(zap.ErrorLevel, "rpc.poolIdempotencyDone", "json.Unmarshal", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return false, err
	}
	return entry.Done, nil
}
func callPoolHandler(ctx context.Context, req interface{}, handler grpc.UnaryHandler) (response interface{}, err error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(req)...)
	poolLogTrace := poollog.Begin("rpc.callPoolHandler", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&err)
	defer func() {
		if recovered := recover(); recovered != nil {
			poollog.Event(zap.ErrorLevel, "rpc.callPoolHandler", "handler_panic", "开奖资金处理发生异常，保留原错误恢复行为", append(poollog.MetadataFields(req), zap.String("panicType", fmt.Sprintf("%T", recovered)), zap.Stack("stack"))...)
			err = status.Error(codes.Internal, fmt.Sprintf("水池资金操作结果待核实：%v", recovered))
		}
	}()
	return handler(ctx, req)
}
