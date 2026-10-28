package rpc

import (
	"app/config"
	"app/poolgroup"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"lottery/dao"
	"micro_service/services"
	"strconv"
	"strings"
	"time"
)

type poolRequestIdentity interface {
	GetAgentId() uint32
	GetGameId() uint32
}
type poolResponseCode interface{ GetCode() services.ErrorCode }
type sessionChange struct {
	Key   string
	Value *dao.PoolSession
}

func poolRequestSession(req interface{}) (key string, newBet bool) {
	switch r := req.(type) {
	case *services.SlotsDoBetReq:
		return dao.PoolSessionKey("slots", r.UserId, r.GameId, r.CurrencyType, ""), true
	case *services.SlotsDoBetFreeReq:
		return dao.PoolSessionKey("slots", r.UserId, r.GameId, r.CurrencyType, ""), true
	case *services.Ug5GuessReq:
		return dao.PoolSessionKey("ug5", r.UserId, r.GameId, r.CurrencyType, ""), true
	case *services.FruitDoBetReq:
		return dao.PoolSessionKey("fruit", r.UserId, r.GameId, r.CurrencyType, ""), true
	case *services.Luck5DealReq:
		return dao.Luck5SessionKey(r.UserId, r.GameId, r.CurrencyType, r.RoundId), true
	case *services.Luck5SettleReq:
		return dao.Luck5SessionKey(r.UserId, r.GameId, r.CurrencyType, r.RoundId), false
	case *services.FruitDoBetMultiReq:
		return dao.PoolSessionKey("multi", r.UserId, r.GameId, r.CurrencyType, ""), true
	case *services.FruitRefundMultiReq:
		return dao.PoolSessionKey("multi", r.UserId, r.GameId, r.CurrencyType, ""), false
	case *services.FruitSettleRoundReq:
		return "", false
	}
	return "", false
}

// poolGroupInterceptor serializes each agent across lottery instances and refreshes the applied mapping on every call.
func poolGroupInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	identity, ok := req.(poolRequestIdentity)
	if !ok {
		return handler(ctx, req)
	}
	agentID := int64(identity.GetAgentId())
	rdb := dao.PoolGroupStore().Redis
	acquireCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	lease, err := poolgroup.Acquire(acquireCtx, rdb, agentID)
	if err != nil {
		return nil, status.Error(codes.Unavailable, "水池正忙，请重试")
	}
	defer lease.Close()
	if err = dao.BeginPoolRequest(ctx, agentID); err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}
	if _, ok := req.(*services.PoolAmountResultReq); ok {
		return callPoolHandler(ctx, req, handler)
	}
	game := dao.GamesManagerIns().GetById(int64(identity.GetGameId()))
	if game == nil || game.ConfName == "" {
		return handler(ctx, req)
	}
	sessionKey, newBet := poolRequestSession(req)
	sessionsKey := poolgroup.Key(agentID, "sessions")
	oldRaw, err := rdb.HGet(ctx, sessionsKey, sessionKey).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, status.Error(codes.Unavailable, "读取活动订单失败")
	}
	var old *dao.PoolSession
	if len(oldRaw) > 0 {
		old = &dao.PoolSession{}
		if json.Unmarshal(oldRaw, old) != nil {
			return nil, status.Error(codes.Unavailable, "活动订单数据无效")
		}
	}
	if old != nil && old.PoolSymbol != "" && old.PoolSymbol != config.CfgIns.PoolStorageSymbol(agentID, game.ConfName) {
		return nil, status.Error(codes.FailedPrecondition, "未结算订单的原水池与当前配置不一致")
	}
	// Pending operations pause new bets but never block a recorded continuation or settlement.
	pendingRaw, err := rdb.Get(ctx, poolgroup.Key(agentID, "pending")).Bytes()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, status.Error(codes.Unavailable, "读取水池切换状态失败")
	}
	pending := 0
	if len(pendingRaw) > 0 {
		var gate poolgroup.Gate
		if json.Unmarshal(pendingRaw, &gate) != nil {
			return nil, status.Error(codes.Unavailable, "水池切换状态无效")
		}
		if gate.Affects(game.ConfName) {
			pending = 1
		}
	}
	if pending > 0 && newBet && old == nil {
		return nil, status.Error(codes.Unavailable, "水池切换中，暂不接受新下注")
	}
	if pending > 0 && newBet {
		switch req.(type) {
		case *services.FruitDoBetMultiReq, *services.SlotsDoBetReq:
			return nil, status.Error(codes.Unavailable, "水池切换中，请先完成当前局")
		}
	}
	raw, _ := json.Marshal(req)
	sum := sha256.Sum256(append([]byte(info.FullMethod), raw...))
	fingerprint := hex.EncodeToString(sum[:])
	// These protocols deliberately treat identical packets as separate increments.
	repeatable := false
	switch r := req.(type) {
	case *services.FruitDoBetMultiReq, *services.FruitRefundMultiReq:
		repeatable = true
	case *services.Ug5GuessReq:
		repeatable = !r.Complete
	}
	trackerKey := poolgroup.Key(agentID, "seen:"+fingerprint)
	primaryKey := primaryPoolIdempotencyKey(req)
	if primaryKey != "" {
		digest := sha256.Sum256([]byte(primaryKey))
		trackerKey = poolgroup.Key(agentID, "seen:"+hex.EncodeToString(digest[:]))
	}
	seen, err := rdb.Exists(ctx, trackerKey).Result()
	if err != nil {
		return nil, status.Error(codes.Unavailable, "读取订单跟踪失败")
	}
	primaryDone := false
	if primaryKey != "" {
		primaryDone, err = poolIdempotencyDone(ctx, rdb, primaryKey)
		if err != nil {
			return nil, status.Error(codes.Unavailable, "读取结算幂等状态失败")
		}
		if !primaryDone {
			seen = 0
		}
	}
	token := fingerprint
	if repeatable {
		b := make([]byte, 16)
		if _, err = rand.Read(b); err != nil {
			return nil, err
		}
		token = hex.EncodeToString(b)
	}
	inflightKey := poolgroup.Key(agentID, "inflight")
	uncertain, readErr := rdb.HGetAll(ctx, inflightKey).Result()
	if readErr != nil {
		return nil, status.Error(codes.Unavailable, "读取处理中订单失败")
	}
	for priorToken, priorRaw := range uncertain {
		var previous struct {
			Recoverable bool `json:"recoverable"`
		}
		if json.Unmarshal([]byte(priorRaw), &previous) != nil {
			return nil, status.Error(codes.Unavailable, "处理中订单数据无效")
		}
		durable := false
		switch req.(type) {
		case *services.Luck5DealReq, *services.Luck5SettleReq:
			durable = true
		}
		if priorToken != token || (!previous.Recoverable && !durable) {
			return nil, status.Error(codes.Unavailable, "该代理有结果待核实的订单，暂不能开始其他资金操作")
		}
	}
	// Cached responses predating tracking must not resurrect an already-finished chain.
	if primaryDone && len(uncertain) == 0 {
		seen = 1
	}
	inflight := map[string]interface{}{"method": info.FullMethod, "symbol": game.ConfName, "session": sessionKey, "createdAt": time.Now().Unix(), "request": json.RawMessage(raw)}
	inflightRaw, _ := json.Marshal(inflight)
	if err = rdb.HSet(ctx, inflightKey, token, inflightRaw).Err(); err != nil {
		return nil, status.Error(codes.Unavailable, "登记处理中订单失败")
	}
	response, callErr := callPoolHandler(ctx, req, handler)
	changed := dao.PoolRequestDirty(agentID)
	finishCtx, finishCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer finishCancel()
	if err = dao.FlushPoolRequest(finishCtx, agentID, lease); err != nil {
		return nil, status.Error(codes.Unavailable, "结算持久化结果待核实："+err.Error())
	}
	// Unknown outcomes retain the inflight record and prevent clearing financial data.
	if callErr != nil {
		return response, callErr
	}
	code, ok := response.(poolResponseCode)
	if !ok {
		return nil, status.Error(codes.Internal, "unknown lottery response")
	}
	if code.GetCode() == services.ErrorCode_SYSTEM_ERROR && changed {
		return response, nil
	}
	if code.GetCode() == services.ErrorCode_OK {
		if primaryKey != "" {
			done, checkErr := poolIdempotencyDone(finishCtx, rdb, primaryKey)
			if checkErr != nil || !done {
				return nil, status.Error(codes.Unavailable, "结算完成标记未保存，资金结果需核实")
			}
		}
		inflight["recoverable"] = true
		data, _ := json.Marshal(inflight)
		if err = rdb.HSet(finishCtx, inflightKey, token, data).Err(); err != nil {
			return nil, status.Error(codes.Unavailable, "成功订单跟踪待恢复")
		}
	}
	if code.GetCode() == services.ErrorCode_OK && (repeatable || seen == 0) {
		changes, err := poolSessionChanges(req, response, old, agentID, game.ConfName, rdb)
		if err != nil {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
		updates := make([][]string, 0, len(changes))
		for _, change := range changes {
			data := ""
			if change.Value != nil {
				raw, _ := json.Marshal(change.Value)
				data = string(raw)
			}
			updates = append(updates, []string{change.Key, data})
		}
		data, _ := json.Marshal(updates)
		mark := "0"
		if !repeatable {
			mark = "1"
		}
		switch req.(type) {
		case *services.Luck5DealReq, *services.Luck5SettleReq:
			mark = "2"
		}
		err = finishPoolTracking.Run(finishCtx, rdb, []string{poolgroup.Key(agentID, "lock"), sessionsKey, trackerKey, inflightKey}, lease.Token, string(data), mark, token).Err()
		if err != nil {
			return nil, status.Error(codes.Unavailable, "保存订单归属失败")
		}
	} else if err = finishPoolTracking.Run(finishCtx, rdb, []string{poolgroup.Key(agentID, "lock"), sessionsKey, trackerKey, inflightKey}, lease.Token, "[]", "0", token).Err(); err != nil {
		return nil, status.Error(codes.Unavailable, "完成订单跟踪失败")
	}
	return response, nil
}

var finishPoolTracking = redis.NewScript(`
if redis.call('GET',KEYS[1])~=ARGV[1] then return redis.error_reply('pool lease lost') end
for _,change in ipairs(cjson.decode(ARGV[2])) do
 if change[2]=='' then redis.call('HDEL',KEYS[2],change[1]) else redis.call('HSET',KEYS[2],change[1],change[2]) end
end
if ARGV[3]=='1' then redis.call('SET',KEYS[3],'1','EX',7200) end
if ARGV[3]=='2' then redis.call('SET',KEYS[3],'1') end
redis.call('HDEL',KEYS[4],ARGV[4])
return 1
`)

func primaryPoolIdempotencyKey(req interface{}) string {
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
func poolIdempotencyDone(ctx context.Context, rdb redis.UniversalClient, key string) (bool, error) {
	raw, err := rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var entry idempotencyEntry
	if err = json.Unmarshal(raw, &entry); err != nil {
		return false, err
	}
	return entry.Done, nil
}

func callPoolHandler(ctx context.Context, req interface{}, handler grpc.UnaryHandler) (response interface{}, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = status.Error(codes.Internal, fmt.Sprintf("水池资金操作结果待核实：%v", recovered))
		}
	}()
	return handler(ctx, req)
}

func poolSessionChanges(req, response interface{}, old *dao.PoolSession, agentID int64, symbol string, rdb redis.UniversalClient) ([]sessionChange, error) {
	key, _ := poolRequestSession(req)
	session := &dao.PoolSession{Symbol: symbol, PoolSymbol: config.CfgIns.PoolStorageSymbol(agentID, symbol), Amount: "0"}
	if group := config.CfgIns.GetPoolGroup(agentID, symbol); group != nil {
		session.Version = group.Version
	}
	if old != nil {
		copy := *old
		session = &copy
	}
	set := func(keep bool) []sessionChange {
		if keep {
			return []sessionChange{{Key: key, Value: session}}
		}
		return []sessionChange{{Key: key}}
	}
	switch r := req.(type) {
	case *services.SlotsDoBetReq:
		session.Kind = "slots"
		session.UserID = r.UserId
		session.RoundID = r.RoundId
		refund := "0"
		if r.FlowControl != nil && r.FlowControl.RefundPoolAmount != "" {
			refund = r.FlowControl.RefundPoolAmount
		}
		amount, keep := dao.RemainingPoolSession(session.Amount, r.PreWinAmount, refund)
		session.Amount = amount
		return set(keep), nil
	case *services.SlotsDoBetFreeReq:
		refund := parseAmount(r.WinAmount)
		if r.FlowControl != nil {
			refund = refund.Add(parseAmount(r.FlowControl.RefundPoolAmount))
		}
		amount, keep := dao.RemainingPoolSession(session.Amount, "0", refund.String())
		session.Amount = amount
		return set(old != nil && keep), nil
	case *services.FruitDoBetReq:
		session.Kind = "fruit"
		session.UserID = r.UserId
		session.RoundID = r.RoundId
		return set(!r.Complete), nil
	case *services.Ug5GuessReq:
		session.Kind = "ug5"
		session.UserID = r.UserId
		session.RoundID = r.RoundId
		return set(!r.Complete), nil
	case *services.Luck5DealReq:
		session.Kind = "luck5"
		session.UserID = r.UserId
		session.RoundID = r.RoundId
		return set(true), nil
	case *services.Luck5SettleReq:
		return set(false), nil
	case *services.FruitDoBetMultiReq:
		session.Kind = "multi"
		session.UserID = r.UserId
		amount, keep := dao.RemainingPoolSession(session.Amount, r.BetAmount, "0")
		session.Amount = amount
		return set(keep), nil
	case *services.FruitRefundMultiReq:
		amount, keep := dao.RemainingPoolSession(session.Amount, "0", r.BetAmount)
		session.Amount = amount
		return set(old != nil && keep), nil
	case *services.FruitSettleRoundReq:
		changes := make([]sessionChange, 0, len(r.Players))
		resp := response.(*services.FruitSettleRoundResp)
		settled := map[uint32]bool{}
		for _, p := range resp.Players {
			settled[p.UserId] = true
		}
		for _, p := range r.Players {
			if !settled[p.UserId] {
				continue
			}
			k := dao.PoolSessionKey("multi", p.UserId, r.GameId, r.CurrencyType, "")
			raw, err := rdb.HGet(context.Background(), poolgroup.Key(agentID, "sessions"), k).Bytes()
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return nil, err
			}
			var existing dao.PoolSession
			if err = json.Unmarshal(raw, &existing); err != nil {
				return nil, err
			}
			amount, keep := dao.RemainingPoolSession(existing.Amount, "0", p.TotalBet)
			existing.Amount = amount
			if keep {
				changes = append(changes, sessionChange{Key: k, Value: &existing})
			} else {
				changes = append(changes, sessionChange{Key: k})
			}
		}
		return changes, nil
	}
	return nil, fmt.Errorf("untracked pool method: %s", strconv.FormatInt(agentID, 10)+strings.TrimSpace(symbol))
}
