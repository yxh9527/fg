package dao

import (
	"app/config"
	"app/poolgroup"
	"app/poollog"
	"app/tables/manager"
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func PoolGroupStore() (poolLogResult0 poolgroup.Store) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("dao.PoolGroupStore", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return poolgroup.Store{DB: MDIns(), Redis: RedisIns().cli}
}

var errPoolResetWaiting = poollog.Expected(errors.New("等待该代理前序水池配置任务完成"))

// BeginPoolRequest runs while the cross-instance agent lease is held. Never reuse another instance's cached totals.
func BeginPoolRequest(ctx context.Context, agentID int64) (poolLogResult0 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogTrace := poollog.Begin("dao.BeginPoolRequest", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult0)
	state, err := PoolGroupStore().Load(ctx, agentID)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.BeginPoolRequest", "PoolGroupStore.Load", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	config.CfgIns.ReplacePoolGroups(agentID, state.Groups)
	agent := CacheIns().GetAgent(agentID)
	agent.lock.Lock()
	defer agent.lock.Unlock()
	agent.poolManaged = true
	agent.poolRevision = state.Revision
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

func FlushPoolRequest(ctx context.Context, agentID int64, lease *poolgroup.Lease) (poolLogResult0 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogTrace := poollog.Begin("dao.FlushPoolRequest", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult0)
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
		poollog.Event(zap.ErrorLevel, "dao.FlushPoolRequest", "flushPoolRequestScript.Run.Err", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	for _, g := range agent.gameCache {
		g.UpdateTime = 0
	}
	for _, u := range agent.userCache {
		u.UpdateTime = 0
	}
	poollog.Event(zap.DebugLevel, "dao.FlushPoolRequest", "redis_flush_done", "水池及玩家统计已写入Redis", append(poolLogFields, zap.Int("poolCount", len(games)), zap.Int("userCount", len(users)))...)
	return nil
}
func PoolRequestDirty(agentID int64) (poolLogResult0 bool) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogTrace := poollog.BeginDebug("dao.PoolRequestDirty", poolLogFields...)
	defer poolLogTrace.Finish(nil)
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
func StartPoolGroupWorker() (poolLogResult0 error) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.Begin("dao.StartPoolGroupWorker", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult0)
	if err := poolgroup.Migrate(MDIns()); err != nil {
		poollog.Event(zap.ErrorLevel, "dao.StartPoolGroupWorker", "poolgroup.Migrate", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	go func() {
		poollog.Event(zap.InfoLevel, "dao.StartPoolGroupWorker", "worker_started", "水池任务执行器已启动，轮询间隔1秒")
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var ops []manager.PoolGroupOperation
			first := MDIns().Model(&manager.PoolGroupOperation{}).Select("MIN(id)").Where("status <> ? AND kind <> ?", poolgroup.Applied, "reset").Group("agent_id")
			if err := MDIns().Where("id IN (?)", first).Order("updated_at ASC, id ASC").Limit(100).Find(&ops).Error; err != nil {
				poollog.Event(zap.ErrorLevel, "dao.StartPoolGroupWorker", "MDIns.Where.Order.Limit.Find.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				zap.L().Error("读取水池组任务失败", zap.Error(err))
				continue
			}
			seen := map[int64]bool{}
			if len(ops) > 0 {
				poollog.Event(zap.DebugLevel, "dao.StartPoolGroupWorker", "tasks_loaded", "读取到待处理水池任务", zap.Int("operationCount", len(ops)))
			}
			for _, op := range ops {
				if seen[op.AgentID] {
					continue
				}
				seen[op.AgentID] = true
				func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if err := executePoolGroupOperation(ctx, op); err != nil && !errors.Is(err, errPoolResetWaiting) {
						if err != nil {
							poollog.Event(zap.ErrorLevel, "dao.StartPoolGroupWorker", "executePoolGroupOperation", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
						}
						if saveErr := MDIns().Model(&manager.PoolGroupOperation{}).Where("id = ? AND status <> ?", op.ID, poolgroup.Applied).Updates(map[string]interface{}{"status": poolgroup.Failed, "error": err.Error()}).Error; saveErr != nil {
							poollog.Event(zap.ErrorLevel, "dao.StartPoolGroupWorker", "failure_status_write_failed", "执行失败且写回任务失败状态也失败", append(poollog.MetadataFields(op), zap.Error(saveErr))...)
						}
						zap.L().Error("水池组任务执行失败，将重试", zap.Int64("operationId", op.ID), zap.Error(err))
					}
				}()
			}
		}
	}()
	return nil
}
func executePoolGroupOperation(ctx context.Context, op manager.PoolGroupOperation) (poolLogResult0 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(op)...)
	poolLogTrace := poollog.Begin("dao.executePoolGroupOperation", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult0)
	lease, err := poolgroup.Acquire(ctx, RedisIns().cli, op.AgentID)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "poolgroup.Acquire", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	defer lease.Close()
	var current manager.PoolGroupOperation
	if err := MDIns().First(&current, op.ID).Error; err != nil {
		poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "MDIns.First.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	if current.Status == poolgroup.Applied {
		poollog.Event(zap.InfoLevel, "dao.executePoolGroupOperation", "already_applied", "任务已执行，仅检查并同步本机缓存", poolLogFields...)
		state, err := PoolGroupStore().Load(ctx, op.AgentID)
		if err != nil {
			poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "PoolGroupStore.Load", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return err
		}
		syncPoolRuntimeMemory(op.AgentID, state)
		return nil
	}
	op = current
	var earlier int64
	if err = MDIns().WithContext(ctx).Model(&manager.PoolGroupOperation{}).Where("agent_id = ? AND id < ? AND status <> ? AND kind <> ?", op.AgentID, op.ID, poolgroup.Applied, "reset").Count(&earlier).Error; err != nil {
		poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "MDIns.WithContext.Model.Where.Count.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	if earlier > 0 {
		poollog.Event(zap.WarnLevel, "dao.executePoolGroupOperation", "waiting_earlier_tasks", "存在未完成前序任务，本任务暂不执行", append(poolLogFields, zap.Int64("earlierCount", earlier))...)
		return errPoolResetWaiting
	}
	var plan poolgroup.Plan
	if err = json.Unmarshal([]byte(op.Payload), &plan); err != nil {
		poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "json.Unmarshal", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	state, err := PoolGroupStore().Load(ctx, op.AgentID)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "PoolGroupStore.Load", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	gate := poolgroup.OperationGate(state, op, plan)
	gateRaw, _ := json.Marshal(gate)
	if err = RedisIns().cli.Set(ctx, poolgroup.Key(op.AgentID, "pending"), gateRaw, 0).Err(); err != nil {
		poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "RedisIns.cli.Set.Err", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	if state.LastOperationID < op.ID {
		poollog.Event(zap.InfoLevel, "dao.executePoolGroupOperation", "gate_enabled", "已开启受影响游戏的切换门禁", append(poolLogFields, zap.Bool("allGames", gate.All), zap.Int("affectedGameCount", len(gate.Symbols)))...)
		if err = MDIns().Model(&manager.PoolGroupOperation{}).Where("id = ?", op.ID).Updates(map[string]interface{}{"status": poolgroup.Applying, "error": ""}).Error; err != nil {
			poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "MDIns.Model.Where.Updates.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return err
		}
		poollog.Event(zap.InfoLevel, "dao.executePoolGroupOperation", "status_applying", "任务进入实际执行阶段", append(poolLogFields, zap.String("status", poolgroup.Applying))...)
		state, err = applyPoolRuntime(ctx, lease, state, op, plan)
		if err != nil {
			poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "applyPoolRuntime", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return err
		}
	}
	syncPoolRuntimeMemory(op.AgentID, state)
	// Compare the operation number: a retry must never remove another operation's gate.
	if err = clearPoolGate.Run(ctx, RedisIns().cli, []string{poolgroup.Key(op.AgentID, "pending"), lease.Key}, string(gateRaw), lease.Token).Err(); err != nil {
		poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "clearPoolGate.Run.Err", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	// A SQL retry sees the already-applied Redis operation and cannot reset subsequent bets.
	poollog.Event(zap.InfoLevel, "dao.executePoolGroupOperation", "gate_cleared", "切换门禁已解除，准备确认数据库任务", poolLogFields...)
	if err := PoolGroupStore().Finalize(ctx, op, plan); err != nil {
		poollog.Event(zap.ErrorLevel, "dao.executePoolGroupOperation", "PoolGroupStore.Finalize", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	if op.Kind == "reset" {
		zap.L().Info("代理水池重置完成，Redis统计和内存缓存已清理", zap.Int64("agentId", op.AgentID), zap.Int64("operationId", op.ID))
	}
	poollog.Event(zap.InfoLevel, "dao.executePoolGroupOperation", "mysql_finalized", "水池任务已在数据库确认完成", append(poolLogFields, zap.String("status", poolgroup.Applied), zap.Int64("runtimeRevision", state.Revision))...)
	return nil
}

// applyPoolRuntime clears Redis first; only a confirmed write may invalidate live cache data.
func applyPoolRuntime(ctx context.Context, lease *poolgroup.Lease, state *poolgroup.State, op manager.PoolGroupOperation, plan poolgroup.Plan) (poolLogResult0 *poolgroup.State, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(state)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(op)...)
	poolLogTrace := poollog.Begin("dao.applyPoolRuntime", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	next, err := PoolGroupStore().Apply(ctx, lease, state, op, plan)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.applyPoolRuntime", "PoolGroupStore.Apply", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	syncPoolRuntimeMemory(op.AgentID, next)
	return next, nil
}

// Every subscriber invalidates its own cache, even if another instance applied the batch.
func syncPoolRuntimeMemory(agentID int64, state *poolgroup.State) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogFields = append(poolLogFields, poollog.MetadataFields(state)...)
	poolLogTrace := poollog.Begin("dao.syncPoolRuntimeMemory", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(nil)
	config.CfgIns.ReplacePoolGroups(agentID, state.Groups)
	agent := CacheIns().GetAgent(agentID)
	agent.lock.Lock()
	defer agent.lock.Unlock()
	if agent.poolManaged && agent.poolRevision == state.Revision {
		poollog.Event(zap.DebugLevel, "dao.syncPoolRuntimeMemory", "cache_current", "本机水池缓存已是当前版本，保留新投注", poolLogFields...)
		return
	}
	cacheCount := len(agent.gameCache)
	previousRevision := agent.poolRevision
	for _, game := range agent.gameCache {
		game.TotalChips = decimal.Zero
		game.TotalEffectBet = decimal.Zero
		game.TotalProfLoss = decimal.Zero
		game.TotalRevenue = decimal.Zero
		game.UpdateTime = 0
	}
	agent.poolManaged = true
	agent.poolRevision = state.Revision
	agent.gameCache = make(map[string]*Game)
	poollog.Event(zap.InfoLevel, "dao.syncPoolRuntimeMemory", "memory_invalidated", "本机旧水池缓存和待写回标记已清理", append(poolLogFields, zap.Int("cacheCount", cacheCount), zap.Int64("previousRevision", previousRevision), zap.Int64("currentRevision", state.Revision))...)
}

var clearPoolGate = redis.NewScript(`if redis.call('GET',KEYS[2])~=ARGV[2] then return redis.error_reply('pool lease lost') end if redis.call('GET',KEYS[1])==ARGV[1] then redis.call('DEL',KEYS[1]) end return 1`)
