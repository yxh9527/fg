package dao

import (
	"app/config"
	"app/entity"
	"app/poolgroup"
	"app/tables/manager"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func PoolGroupStore() poolgroup.Store { return poolgroup.Store{DB: MDIns(), Redis: RedisIns().cli} }

func queuePoolReset(data interface{}) {
	reset, ok := data.(*entity.ResetPool)
	if !ok || reset.BatchID == "" {
		zap.L().Error("重置任务缺少批次编号，拒绝非幂等清池")
		return
	}
	if err := PoolGroupStore().QueueReset(context.Background(), reset.BatchID); err != nil {
		zap.L().Error("创建水池重置任务失败", zap.Error(err))
	}
}

// BeginPoolRequest runs while the cross-instance agent lease is held. Never reuse another instance's cached totals.
func BeginPoolRequest(ctx context.Context, agentID int64) error {
	state, err := PoolGroupStore().Load(ctx, agentID)
	if err != nil {
		return err
	}
	config.CfgIns.ReplacePoolGroups(agentID, state.Groups)
	agent := CacheIns().GetAgent(agentID)
	agent.lock.Lock()
	defer agent.lock.Unlock()
	agent.poolManaged = true
	agent.gameCache = make(map[string]*Game)
	agent.userCache = make(map[uint32]*User)
	return nil
}

var flushPoolRequestScript = redis.NewScript(`
if redis.call('GET',KEYS[1])~=ARGV[1] then return redis.error_reply('pool lease lost') end
for _,g in ipairs(cjson.decode(ARGV[2])) do
 for i=1,4 do redis.call('ZADD',KEYS[i+1],g[i+1],g[1]) end
end
for _,u in ipairs(cjson.decode(ARGV[3])) do
 redis.call('ZADD','userTotalEffBet',u[2],u[1])
 redis.call('ZADD','userTotalProfLoss',u[3],u[1])
 redis.call('ZADD','userBetCount',u[4],u[1])
end
return 1
`)

func FlushPoolRequest(ctx context.Context, agentID int64, lease *poolgroup.Lease) error {
	agent := CacheIns().GetAgent(agentID)
	agent.lock.Lock()
	defer agent.lock.Unlock()
	games := make([][]string, 0, len(agent.gameCache))
	users := make([][]string, 0, len(agent.userCache))
	for symbol, g := range agent.gameCache {
		if g.UpdateTime > 0 {
			games = append(games, []string{poolgroup.Member(agentID, symbol), g.TotalEffectBet.String(), g.TotalChips.String(), g.TotalProfLoss.String(), g.TotalRevenue.String()})
		}
	}
	for id, u := range agent.userCache {
		if u.UpdateTime > 0 {
			users = append(users, []string{strconv.FormatUint(uint64(id), 10), u.TotalEffectBet.String(), u.TotalProfLoss.String(), u.Count.String()})
		}
	}
	graw, _ := json.Marshal(games)
	uraw, _ := json.Marshal(users)
	keys := append([]string{poolgroup.Key(agentID, "lock")}, poolgroup.StatKeys()...)
	if err := flushPoolRequestScript.Run(ctx, RedisIns().cli, keys, lease.Token, string(graw), string(uraw)).Err(); err != nil {
		return err
	}
	for _, g := range agent.gameCache {
		g.UpdateTime = 0
	}
	for _, u := range agent.userCache {
		u.UpdateTime = 0
	}
	return nil
}

func PoolRequestDirty(agentID int64) bool {
	agent := CacheIns().GetAgent(agentID)
	agent.lock.RLock()
	defer agent.lock.RUnlock()
	for _, g := range agent.gameCache {
		if g.UpdateTime > 0 {
			return true
		}
	}
	for _, u := range agent.userCache {
		if u.UpdateTime > 0 {
			return true
		}
	}
	return false
}

// StartPoolGroupWorker polls SQL tasks, so lost Pub/Sub messages cannot lose an operation.
func StartPoolGroupWorker() error {
	if err := poolgroup.Migrate(MDIns()); err != nil {
		return err
	}
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var ops []manager.PoolGroupOperation
			first := MDIns().Model(&manager.PoolGroupOperation{}).Select("MIN(id)").Where("status <> ?", poolgroup.Applied).Group("agent_id")
			if err := MDIns().Where("id IN (?)", first).Order("updated_at ASC, id ASC").Limit(100).Find(&ops).Error; err != nil {
				zap.L().Error("读取水池组任务失败", zap.Error(err))
				continue
			}
			seen := map[int64]bool{}
			for _, op := range ops {
				if seen[op.AgentID] {
					continue
				}
				seen[op.AgentID] = true
				func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := executePoolGroupOperation(ctx, op); err != nil {
						_ = MDIns().Model(&manager.PoolGroupOperation{}).Where("id = ? AND status <> ?", op.ID, poolgroup.Applied).Updates(map[string]interface{}{"status": poolgroup.Failed, "error": err.Error()}).Error
						zap.L().Error("水池组任务执行失败，将重试", zap.Int64("operationId", op.ID), zap.Error(err))
					}
				}()
			}
		}
	}()
	return nil
}

func executePoolGroupOperation(ctx context.Context, op manager.PoolGroupOperation) error {
	lease, err := poolgroup.Acquire(ctx, RedisIns().cli, op.AgentID)
	if err != nil {
		return err
	}
	defer lease.Close()
	var current manager.PoolGroupOperation
	if err := MDIns().First(&current, op.ID).Error; err != nil {
		return err
	}
	if current.Status == poolgroup.Applied {
		return nil
	}
	op = current
	var plan poolgroup.Plan
	if err = json.Unmarshal([]byte(op.Payload), &plan); err != nil {
		return err
	}
	state, err := PoolGroupStore().Load(ctx, op.AgentID)
	if err != nil {
		return err
	}
	gate := poolgroup.OperationGate(state, op, plan)
	gateRaw, _ := json.Marshal(gate)
	if err = RedisIns().cli.Set(ctx, poolgroup.Key(op.AgentID, "pending"), gateRaw, 0).Err(); err != nil {
		return err
	}
	if state.LastOperationID < op.ID {
		sessions, err := RedisIns().cli.HGetAll(ctx, poolgroup.Key(op.AgentID, "sessions")).Result()
		if err != nil {
			return err
		}
		processing, err := RedisIns().cli.HGetAll(ctx, poolgroup.Key(op.AgentID, "inflight")).Result()
		if err != nil {
			return err
		}
		open, inflight := 0, 0
		for _, raw := range sessions {
			var session PoolSession
			if err = json.Unmarshal([]byte(raw), &session); err != nil {
				return err
			}
			if gate.Affects(session.Symbol) {
				open++
			}
		}
		for _, raw := range processing {
			var request struct {
				Symbol string `json:"symbol"`
			}
			if err = json.Unmarshal([]byte(raw), &request); err != nil {
				return err
			}
			if gate.Affects(request.Symbol) {
				inflight++
			}
		}
		agent := CacheIns().GetAgent(op.AgentID)
		agent.lock.Lock()
		legacy := false
		for _, r := range agent.RoundCache {
			if !r.Over {
				legacy = true
				break
			}
		}
		agent.lock.Unlock()
		if open > 0 || inflight > 0 || legacy {
			return MDIns().Model(&manager.PoolGroupOperation{}).Where("id = ?", op.ID).Updates(map[string]interface{}{"status": poolgroup.Waiting, "error": fmt.Sprintf("等待未结算订单：活动链%d，处理中%d", open, inflight)}).Error
		}
		if err = MDIns().Model(&manager.PoolGroupOperation{}).Where("id = ?", op.ID).Updates(map[string]interface{}{"status": poolgroup.Applying, "error": ""}).Error; err != nil {
			return err
		}
		state, err = PoolGroupStore().Apply(ctx, lease, state, op, plan)
		if err != nil {
			return err
		}
	}
	config.CfgIns.ReplacePoolGroups(op.AgentID, state.Groups)
	agent := CacheIns().GetAgent(op.AgentID)
	agent.lock.Lock()
	agent.poolManaged = true
	agent.gameCache = make(map[string]*Game)
	agent.lock.Unlock()
	// Compare the operation number: a retry must never remove another operation's gate.
	if err = clearPoolGate.Run(ctx, RedisIns().cli, []string{poolgroup.Key(op.AgentID, "pending"), lease.Key}, string(gateRaw), lease.Token).Err(); err != nil {
		return err
	}
	// A SQL retry sees the already-applied Redis operation and cannot reset subsequent bets.
	return PoolGroupStore().Finalize(ctx, op, plan)
}

var clearPoolGate = redis.NewScript(`if redis.call('GET',KEYS[2])~=ARGV[2] then return redis.error_reply('pool lease lost') end if redis.call('GET',KEYS[1])==ARGV[1] then redis.call('DEL',KEYS[1]) end return 1`)

// RecoverLuck5Sessions imports durable accepted deals made before an interceptor/session update.
func RecoverLuck5Sessions(ctx context.Context) error {
	var cursor uint64
	for {
		keys, next, err := RedisIns().cli.Scan(ctx, cursor, "luck5:round:*", 100).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			raw, err := RedisIns().cli.Get(ctx, key).Bytes()
			if errors.Is(err, redis.Nil) {
				continue
			}
			if err != nil {
				return err
			}
			var r Luck5Round
			if err = json.Unmarshal(raw, &r); err != nil {
				return err
			}
			if r.DealCode == 0 && !r.RecordDone && r.IsTourist == 0 {
				if err = recoverLuck5Session(ctx, key, int64(r.AgentId)); err != nil {
					return err
				}
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}

func recoverLuck5Session(ctx context.Context, key string, agentID int64) error {
	lease, err := poolgroup.Acquire(ctx, RedisIns().cli, agentID)
	if err != nil {
		return err
	}
	defer lease.Close()
	raw, err := RedisIns().cli.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil
	}
	if err != nil {
		return err
	}
	var r Luck5Round
	if err = json.Unmarshal(raw, &r); err != nil {
		return err
	}
	if r.DealCode != 0 || r.RecordDone || r.IsTourist != 0 {
		return nil
	}
	symbol := r.PoolSymbol
	if symbol == "" {
		symbol = r.Symbol
	}
	session := PoolSession{Kind: "luck5", Symbol: r.Symbol, PoolSymbol: symbol, Version: r.PoolVersion, RoundID: r.RoundId, UserID: r.UserId, Amount: "0"}
	data, _ := json.Marshal(session)
	return RedisIns().cli.HSet(ctx, poolgroup.Key(agentID, "sessions"), Luck5SessionKey(r.UserId, r.GameId, r.CurrencyType, r.RoundId), data).Err()
}
