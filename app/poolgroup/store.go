// Package poolgroup coordinates durable administration tasks with lottery execution.
package poolgroup

import (
	"app/config"
	"app/esindex"
	"app/poollog"
	"app/tables/manager"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	Pending  = "pending"
	Waiting  = "waiting"
	Applying = "applying"
	Applied  = "applied"
	Failed   = "failed"
)

type Store struct {
	DB    *gorm.DB
	Redis redis.UniversalClient
}
type Intent struct {
	AgentID   int64    `json:"agentId"`
	GroupID   int64    `json:"groupId"`
	Version   int64    `json:"version"`
	RequestID string   `json:"requestId"`
	Kind      string   `json:"kind"`
	Name      string   `json:"name"`
	Symbols   []string `json:"symbols"`
}
type Plan struct {
	Intent Intent                    `json:"intent"`
	Target *config.PoolGroupSnapshot `json:"target,omitempty"`
}
type State struct {
	Revision        int64                       `json:"revision"`
	LastOperationID int64                       `json:"lastOperationId"`
	Groups          []*config.PoolGroupSnapshot `json:"groups"`
}
type Gate struct {
	OperationID int64    `json:"operationId"`
	Symbols     []string `json:"symbols"`
	All         bool     `json:"all"`
}

func (g Gate) Affects(symbol string) (poolLogResult0 bool) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(g)...)
	poolLogFields = append(poolLogFields, zap.String("symbol", symbol))
	poolLogTrace := poollog.BeginDebug("poolgroup.Gate.Affects", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return g.All || contains(g.Symbols, symbol)
}
func OperationGate(state *State, op manager.PoolGroupOperation, plan Plan) (poolLogResult0 Gate) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(state)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(op)...)
	poolLogTrace := poollog.BeginDebug("poolgroup.OperationGate", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	gate := Gate{OperationID: op.ID, All: op.Kind == "reset"}
	seen := map[string]bool{}
	for _, group := range state.Groups {
		if group.ID == op.GroupID {
			for _, symbol := range group.Symbols {
				seen[symbol] = true
			}
		}
	}
	if plan.Target != nil {
		for _, symbol := range plan.Target.Symbols {
			seen[symbol] = true
		}
	}
	for symbol := range seen {
		gate.Symbols = append(gate.Symbols, symbol)
	}
	return gate
}
func Key(agentID int64, part string) (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogFields = append(poolLogFields, zap.String("keyPart", part))
	poolLogTrace := poollog.BeginDebug("poolgroup.Key", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return fmt.Sprintf("%s:%d:%s", esindex.Name("pool_group"), agentID, part)
}
func StatKeys() (poolLogResult0 []string) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("poolgroup.StatKeys", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return []string{esindex.AgentEffectData(), esindex.AgentChipsData(), esindex.AgentProfitLossData(), esindex.AgentRevenueData()}
}
func Member(agentID int64, symbol string) (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogFields = append(poolLogFields, zap.String("symbol", symbol))
	poolLogTrace := poollog.BeginDebug("poolgroup.Member", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return fmt.Sprintf("%d_%s", agentID, symbol)
}
func Migrate(db *gorm.DB) (poolLogResult0 error) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.Begin("poolgroup.Migrate", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult0)
	return db.AutoMigrate(&manager.PoolGroup{}, &manager.PoolGroupMember{}, &manager.PoolGroupOperation{})
}
func (s Store) Load(ctx context.Context, agentID int64) (poolLogResult0 *State, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(s)...)
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogTrace := poollog.Begin("poolgroup.Store.Load", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	raw, err := s.Redis.Get(ctx, Key(agentID, "state")).Bytes()
	if errors.Is(err, redis.Nil) {
		var count int64
		if err = s.DB.WithContext(ctx).Model(&manager.PoolGroup{}).Where("agent_id = ? AND deleted = ?", agentID, false).Count(&count).Error; err != nil {
			poollog.Event(zap.ErrorLevel, "poolgroup.Store.Load", "s.DB.WithContext.Model.Where.Count.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return nil, err
		}
		if count != 0 {
			return nil, errors.New("水池组运行配置缺失，请先恢复运行配置")
		}
		empty := &State{Groups: []*config.PoolGroupSnapshot{}}
		data, _ := json.Marshal(empty)
		created, initErr := s.Redis.SetNX(ctx, Key(agentID, "state"), data, 0).Result()
		if initErr != nil {
			poollog.Event(zap.ErrorLevel, "poolgroup.Store.Load", "s.Redis.SetNX.Result", "水池步骤失败", append(poolLogFields, zap.Error(initErr))...)
			return nil, initErr
		}
		if !created {
			return s.Load(ctx, agentID)
		}
		return empty, nil
	}
	if err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.Store.Load", "s.Redis.Get.Bytes", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	state := &State{}
	if err = json.Unmarshal(raw, state); err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.Store.Load", "json.Unmarshal", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	for _, g := range state.Groups {
		if g == nil || g.AgentID != agentID || g.Config == nil || g.Config.Pool[1] == nil || config.ValidatePoolGroup(g.Name, g.Symbols) != nil {
			return nil, errors.New("水池组运行配置无效")
		}
	}
	return state, nil
}
func (s Store) Submit(ctx context.Context, input Intent, operatorID int64) (poolLogResult0 *manager.PoolGroupOperation, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(s)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(input)...)
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", operatorID))
	poolLogTrace := poollog.Begin("poolgroup.Store.Submit", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	input.Name = strings.TrimSpace(input.Name)
	if input.AgentID <= 0 || input.RequestID == "" || len(input.RequestID) > 80 {
		return nil, errors.New("请选择有效代理并提供请求编号")
	}
	if input.Kind != "create" && input.Kind != "update" && input.Kind != "delete" {
		return nil, errors.New("操作类型无效")
	}
	if input.Kind == "create" && (input.GroupID != 0 || input.Version != 0) {
		return nil, errors.New("新增水池组不允许指定已有组编号或版本")
	}
	if input.Kind != "create" && (input.GroupID <= 0 || input.Version <= 0) {
		return nil, errors.New("请提供有效水池组编号和版本")
	}
	if input.Kind != "delete" {
		if err := config.ValidatePoolGroup(input.Name, input.Symbols); err != nil {
			poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "config.ValidatePoolGroup", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return nil, err
		}
	}
	var result manager.PoolGroupOperation
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var agent manager.Agent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND isDel = 0", input.AgentID).First(&agent).Error; err != nil {
			poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "tx.Clauses.Where.First.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return errors.New("代理不存在")
		}
		var existing manager.PoolGroupOperation
		err := tx.Where("agent_id = ? AND request_id = ?", input.AgentID, input.RequestID).First(&existing).Error
		if err == nil {
			var old Plan
			if json.Unmarshal([]byte(existing.Payload), &old) != nil {
				return errors.New("历史操作数据无效")
			}
			a, _ := json.Marshal(old.Intent)
			b, _ := json.Marshal(input)
			if string(a) != string(b) {
				return errors.New("请求编号已被其他操作使用")
			}
			result = existing
			poollog.Event(zap.InfoLevel, "poolgroup.Store.Submit", "idempotent_hit", "命中已有水池任务，复用操作编号", append(poolLogFields, zap.Int64("operationId", existing.ID), zap.String("status", existing.Status))...)
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		var count int64
		if err := tx.Model(&manager.PoolGroupOperation{}).Where("agent_id = ? AND status <> ?", input.AgentID, Applied).Count(&count).Error; err != nil {
			poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "tx.Model.Where.Count.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return err
		}
		if count > 0 {
			return errors.New("该代理有待完成的水池操作，请等待完成后再修改")
		}
		plan := Plan{Intent: input}
		before := ""
		if input.Kind != "create" {
			var group manager.PoolGroup
			if err := tx.Where("id = ? AND agent_id = ? AND deleted = ?", input.GroupID, input.AgentID, false).First(&group).Error; err != nil {
				poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "tx.Where.First.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return errors.New("水池组不存在")
			}
			if group.Version != input.Version {
				return errors.New("配置已被修改，请刷新后重试")
			}
			previous, err := readSnapshot(tx, group)
			if err != nil {
				poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "readSnapshot", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return err
			}
			b, _ := json.Marshal(previous)
			before = string(b)
			if input.Kind == "update" {
				copy := *previous
				copy.Name = input.Name
				copy.Symbols = input.Symbols
				copy.Version++
				copy.UpdatedAt = time.Now().Unix()
				plan.Target = &copy
			}
		}
		if input.Kind != "delete" {
			var games []manager.Game
			if err := tx.Where("confName IN ?", input.Symbols).Find(&games).Error; err != nil {
				poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "tx.Where.Find.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return err
			}
			found := map[string]bool{}
			for _, g := range games {
				found[g.ConfName] = true
			}
			for _, symbol := range input.Symbols {
				if !found[symbol] {
					return fmt.Errorf("游戏不存在：%s", symbol)
				}
				pc := config.CfgIns.GetIndependentPoolCfg(input.AgentID, symbol)
				if pc == nil || pc.Pool[1] == nil {
					return fmt.Errorf("游戏缺少有效独立水池配置：%s", symbol)
				}
			}
			if err := tx.Model(&manager.PoolGroupMember{}).Where("agent_id = ? AND symbol IN ? AND group_id <> ?", input.AgentID, input.Symbols, input.GroupID).Count(&count).Error; err != nil {
				poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "tx.Model.Where.Count.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return err
			}
			if count > 0 {
				return errors.New("所选游戏已加入其他水池组，请刷新游戏列表")
			}
		}
		if input.Kind == "create" {
			pc := config.CfgIns.GetIndependentPoolCfg(input.AgentID, input.Symbols[0])
			if pc == nil || pc.Pool[1] == nil {
				return errors.New("第一个游戏缺少有效水池配置")
			}
			raw, err := json.Marshal(pc)
			if err != nil {
				poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "json.Marshal", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return err
			}
			var snapshot config.Pool
			if err = json.Unmarshal(raw, &snapshot); err != nil {
				poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "json.Unmarshal", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return err
			}
			row := manager.PoolGroup{AgentID: input.AgentID, Name: input.Name, SourceSymbol: input.Symbols[0], Config: string(raw), Version: 0, Deleted: true}
			if err = tx.Create(&row).Error; err != nil {
				poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "tx.Create.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return err
			}
			result.GroupID = row.ID
			plan.Target = &config.PoolGroupSnapshot{ID: row.ID, AgentID: input.AgentID, Name: input.Name, SourceSymbol: input.Symbols[0], Symbols: input.Symbols, Config: &snapshot, Version: 1, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
		} else {
			result.GroupID = input.GroupID
		}
		payload, err := json.Marshal(plan)
		if err != nil {
			poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "json.Marshal", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return err
		}
		result.AgentID = input.AgentID
		result.RequestID = input.RequestID
		result.Kind = input.Kind
		result.Payload = string(payload)
		result.Before = before
		result.Status = Pending
		result.OperatorID = operatorID
		return tx.Create(&result).Error
	})
	if err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.Store.Submit", "s.DB.WithContext.Transaction", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	poollog.Event(zap.InfoLevel, "poolgroup.Store.Submit", "mysql_committed", "水池配置任务已提交", append(poolLogFields, zap.Int64("operationId", result.ID), zap.Int64("groupId", result.GroupID), zap.String("status", result.Status))...)
	return &result, nil
}
func readSnapshot(tx *gorm.DB, row manager.PoolGroup) (poolLogResult0 *config.PoolGroupSnapshot, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(row)...)
	poolLogTrace := poollog.Begin("poolgroup.readSnapshot", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	var pc config.Pool
	if err := json.Unmarshal([]byte(row.Config), &pc); err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.readSnapshot", "json.Unmarshal", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	var members []manager.PoolGroupMember
	if err := tx.Where("group_id = ? AND agent_id = ?", row.ID, row.AgentID).Order("position ASC").Find(&members).Error; err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.readSnapshot", "tx.Where.Order.Find.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	symbols := make([]string, 0, len(members))
	for _, m := range members {
		symbols = append(symbols, m.Symbol)
	}
	return &config.PoolGroupSnapshot{ID: row.ID, AgentID: row.AgentID, Name: row.Name, SourceSymbol: row.SourceSymbol, Symbols: symbols, Config: &pc, Version: row.Version, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}

// Finalize is restart-safe: runtime application is marked atomically before SQL finalization.
func (s Store) Finalize(ctx context.Context, op manager.PoolGroupOperation, plan Plan) (poolLogResult0 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(s)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(op)...)
	poolLogTrace := poollog.Begin("poolgroup.Store.Finalize", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult0)
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row manager.PoolGroupOperation
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, op.ID).Error; err != nil {
			poollog.Event(zap.ErrorLevel, "poolgroup.Store.Finalize", "tx.Clauses.First.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return err
		}
		if row.Status == Applied {
			return nil
		}
		if op.Kind != "reset" {
			if err := tx.Where("group_id = ? AND agent_id = ?", op.GroupID, op.AgentID).Delete(&manager.PoolGroupMember{}).Error; err != nil {
				poollog.Event(zap.ErrorLevel, "poolgroup.Store.Finalize", "tx.Where.Delete.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return err
			}
			if plan.Target == nil {
				if err := tx.Model(&manager.PoolGroup{}).Where("id = ? AND agent_id = ?", op.GroupID, op.AgentID).Updates(map[string]interface{}{"deleted": true, "version": gorm.Expr("version + 1")}).Error; err != nil {
					poollog.Event(zap.ErrorLevel, "poolgroup.Store.Finalize", "tx.Model.Where.Updates.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
					return err
				}
			} else {
				g := plan.Target
				raw, err := json.Marshal(g.Config)
				if err != nil {
					poollog.Event(zap.ErrorLevel, "poolgroup.Store.Finalize", "json.Marshal", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
					return err
				}
				if err = tx.Model(&manager.PoolGroup{}).Where("id = ? AND agent_id = ?", g.ID, g.AgentID).Updates(map[string]interface{}{"name": g.Name, "source_symbol": g.SourceSymbol, "config": string(raw), "version": g.Version, "deleted": false}).Error; err != nil {
					poollog.Event(zap.ErrorLevel, "poolgroup.Store.Finalize", "tx.Model.Where.Updates.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
					return err
				}
				for i, symbol := range g.Symbols {
					if err = tx.Create(&manager.PoolGroupMember{AgentID: g.AgentID, GroupID: g.ID, Symbol: symbol, Position: i}).Error; err != nil {
						poollog.Event(zap.ErrorLevel, "poolgroup.Store.Finalize", "tx.Create.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
						return err
					}
				}
			}
		}
		return tx.Model(&row).Updates(map[string]interface{}{"status": Applied, "error": ""}).Error
	})
}

// NextState returns only the counters to remove; existing group counters are never copied or split.
func NextState(state *State, op manager.PoolGroupOperation, plan Plan) (poolLogResult0 *State, poolLogResult1 []string, poolLogResult2 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(state)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(op)...)
	poolLogTrace := poollog.Begin("poolgroup.NextState", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult2)
	if state.LastOperationID >= op.ID {
		poollog.Event(zap.InfoLevel, "poolgroup.NextState", "already_applied", "操作已应用，保持现有余额和运行版本", append(poolLogFields, zap.Int64("currentRevision", state.Revision))...)
		return state, nil, nil
	}
	next := &State{Revision: state.Revision + 1, LastOperationID: op.ID, Groups: []*config.PoolGroupSnapshot{}}
	clear := map[string]bool{}
	if op.Kind == "reset" {
		next.Groups = state.Groups
		for _, g := range state.Groups {
			clear[config.PoolGroupSymbol(g.ID)] = true
		}
		for _, symbol := range plan.Intent.Symbols {
			clear[symbol] = true
		}
	} else {
		var previous *config.PoolGroupSnapshot
		for _, g := range state.Groups {
			if g.ID == op.GroupID {
				previous = g
			} else {
				next.Groups = append(next.Groups, g)
			}
		}
		if op.Kind != "create" && (previous == nil || previous.Version != plan.Intent.Version) {
			return nil, nil, errors.New("运行配置版本不一致")
		}
		if previous != nil {
			for _, symbol := range previous.Symbols {
				if plan.Target == nil || !contains(plan.Target.Symbols, symbol) {
					clear[symbol] = true
				}
			}
		}
		if plan.Target != nil {
			if previous == nil {
				clear[config.PoolGroupSymbol(plan.Target.ID)] = true
			}
			for _, symbol := range plan.Target.Symbols {
				if previous == nil || !contains(previous.Symbols, symbol) {
					clear[symbol] = true
				}
			}
			next.Groups = append(next.Groups, plan.Target)
		} else {
			clear[config.PoolGroupSymbol(op.GroupID)] = true
		}
	}
	members := make([]string, 0, len(clear))
	for symbol := range clear {
		members = append(members, Member(op.AgentID, symbol))
	}
	return next, members, nil
}
func contains(items []string, s string) (poolLogResult0 bool) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("poolgroup.contains", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	for _, item := range items {
		if item == s {
			return true
		}
	}
	return false
}

var applyScript = redis.NewScript(`
if redis.call('GET',KEYS[1])~=ARGV[1] then return redis.error_reply('pool lease lost') end
for i=3,6 do local t=redis.call('TYPE',KEYS[i]).ok; if t~='none' and t~='zset' then return redis.error_reply('invalid pool counter type') end end
local current=redis.call('GET',KEYS[2])
if current and tonumber(cjson.decode(current).lastOperationId or 0)>=tonumber(ARGV[3]) then return {0,0} end
local removed=0
for _,member in ipairs(cjson.decode(ARGV[4])) do
 for i=3,6 do removed=removed+redis.call('ZREM',KEYS[i],member) end
end
redis.call('SET',KEYS[2],ARGV[2])
return {1,removed}
`)

func (s Store) Apply(ctx context.Context, lease *Lease, state *State, op manager.PoolGroupOperation, plan Plan) (poolLogResult0 *State, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(s)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(state)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(op)...)
	poolLogTrace := poollog.Begin("poolgroup.Store.Apply", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	next, members, err := NextState(state, op, plan)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.Store.Apply", "NextState", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	if state.LastOperationID >= op.ID {
		return state, nil
	}
	raw, err := json.Marshal(next)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.Store.Apply", "json.Marshal", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	clear, _ := json.Marshal(members)
	keys := append([]string{Key(op.AgentID, "lock"), Key(op.AgentID, "state")}, StatKeys()...)
	poollog.Event(zap.InfoLevel, "poolgroup.Store.Apply", "redis_apply_begin", "准备原子更新水池归属并清理指定统计", append(poolLogFields, zap.Int("clearMemberCount", len(members)), zap.Strings("statKeys", StatKeys()), zap.Int64("nextRevision", next.Revision))...)
	redisResult, err := applyScript.Run(ctx, s.Redis, keys, lease.Token, string(raw), strconv.FormatInt(op.ID, 10), string(clear)).Result()
	if err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.Store.Apply", "applyScript.Run.Err", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	resultFields := append(poolLogFields, zap.Int("clearMemberCount", len(members)), zap.Int64("nextRevision", next.Revision))
	stage, message := "redis_applied", "Redis水池归属及统计清理执行成功"
	if counts, ok := redisResult.([]interface{}); ok && len(counts) == 2 {
		resultFields = append(resultFields, zap.Any("applied", counts[0]), zap.Any("deletedRecordCount", counts[1]))
		if applied, ok := counts[0].(int64); ok && applied == 0 {
			stage, message = "redis_replay_skipped", "Redis已处理相同操作，跳过重复清池"
		}
	} else {
		resultFields = append(resultFields, zap.String("redisResultType", fmt.Sprintf("%T", redisResult)))
	}
	poollog.Event(zap.InfoLevel, "poolgroup.Store.Apply", stage, message, resultFields...)
	return next, nil
}

type Lease struct {
	Redis redis.UniversalClient
	Key   string
	Token string
	stop  chan struct{}
	done  chan struct{}
	Lost  atomic.Bool
}

var renewScript = redis.NewScript(`if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('PEXPIRE',KEYS[1],ARGV[2]) end return 0`)
var releaseScript = redis.NewScript(`if redis.call('GET',KEYS[1])==ARGV[1] then return redis.call('DEL',KEYS[1]) end return 0`)

func Acquire(ctx context.Context, rdb redis.UniversalClient, agentID int64) (poolLogResult0 *Lease, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogTrace := poollog.Begin("poolgroup.Acquire", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	token := make([]byte, 24)
	if _, err := rand.Read(token); err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.Acquire", "rand.Read", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	lease := &Lease{Redis: rdb, Key: Key(agentID, "lock"), Token: hex.EncodeToString(token), stop: make(chan struct{}), done: make(chan struct{})}
	waitStarted := time.Now()
	waitingLogged := false
	for {
		ok, err := rdb.SetNX(ctx, lease.Key, lease.Token, 90*time.Second).Result()
		if err != nil {
			poollog.Event(zap.ErrorLevel, "poolgroup.Acquire", "rdb.SetNX.Result", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return nil, err
		}
		if ok {
			poollog.Event(zap.DebugLevel, "poolgroup.Acquire", "lease_acquired", "已取得代理水池租约", append(poolLogFields, zap.String("lockKey", lease.Key), zap.Duration("waitDuration", time.Since(waitStarted)))...)
			break
		}
		if !waitingLogged {
			poollog.Event(zap.DebugLevel, "poolgroup.Acquire", "lease_wait", "代理水池租约被占用，等待释放", append(poolLogFields, zap.String("lockKey", lease.Key))...)
			waitingLogged = true
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(15 * time.Millisecond):
		}
	}
	go func() {
		defer close(lease.done)
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-lease.stop:
				return
			case <-ticker.C:
				renewCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				n, err := renewScript.Run(renewCtx, rdb, []string{lease.Key}, lease.Token, 90000).Int()
				cancel()
				if err != nil || n != 1 {
					poollog.Event(zap.ErrorLevel, "poolgroup.Acquire", "lease_renew_failed", "代理水池租约续期失败或持有者已变化", append(poolLogFields, zap.String("lockKey", lease.Key), zap.Int("renewResult", n), zap.Error(err))...)
					lease.Lost.Store(true)
					return
				}
			}
		}
	}()
	return lease, nil
}
func (l *Lease) Close() {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(l)...)
	poolLogTrace := poollog.BeginDebug("poolgroup.Lease.Close", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	close(l.stop)
	<-l.done
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := releaseScript.Run(ctx, l.Redis, []string{l.Key}, l.Token).Err(); err != nil {
		poollog.Event(zap.WarnLevel, "poolgroup.Lease.Close", "lease_release_failed", "释放代理水池租约失败，等待租约过期", zap.String("lockKey", l.Key), zap.Error(err))
	} else {
		poollog.Event(zap.DebugLevel, "poolgroup.Lease.Close", "lease_released", "代理水池租约释放检查完成", zap.String("lockKey", l.Key))
	}
}

// ResetOperations loads all batch tasks, including applied tasks for local cache synchronization.
func (s Store) ResetOperations(ctx context.Context, batchID string) (poolLogResult0 []manager.PoolGroupOperation, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(s)...)
	poolLogFields = append(poolLogFields, zap.String("batchId", batchID))
	poolLogTrace := poollog.Begin("poolgroup.Store.ResetOperations", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	var operations []manager.PoolGroupOperation
	err := s.DB.WithContext(ctx).Where("kind = ? AND request_id = ?", "reset", "reset:"+batchID).Order("id ASC").Find(&operations).Error
	if err == nil {
		poollog.Event(zap.InfoLevel, "poolgroup.Store.ResetOperations", "batch_loaded", "已读取重置批次对应任务", append(poolLogFields, zap.Int("operationCount", len(operations)))...)
	}
	return operations, err
}

// QueueReset persists one idempotent reset task per agent; each is drained and executed independently.
func (s Store) QueueReset(ctx context.Context, batchID string) (poolLogResult0 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(s)...)
	poolLogFields = append(poolLogFields, zap.String("batchId", batchID))
	poolLogTrace := poollog.Begin("poolgroup.Store.QueueReset", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult0)
	var agents []manager.Agent
	if err := s.DB.WithContext(ctx).Find(&agents).Error; err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.Store.QueueReset", "s.DB.WithContext.Find.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	agents = append(agents, manager.Agent{Id: 0})
	// A removed agent can still own live Redis statistics and must participate in the reset.
	seen := make(map[int64]bool, len(agents))
	for _, agent := range agents {
		seen[agent.Id] = true
	}
	for _, key := range StatKeys() {
		var cursor uint64
		for {
			values, next, err := s.Redis.ZScan(ctx, key, cursor, "*", 100).Result()
			if err != nil {
				poollog.Event(zap.ErrorLevel, "poolgroup.Store.QueueReset", "s.Redis.ZScan.Result", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return err
			}
			for i := 0; i < len(values); i += 2 {
				separator := strings.IndexByte(values[i], '_')
				if separator < 1 {
					continue
				}
				id, err := strconv.ParseInt(values[i][:separator], 10, 64)
				if err == nil && id >= 0 && !seen[id] {
					agents = append(agents, manager.Agent{Id: id})
					seen[id] = true
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
	var games []manager.Game
	if err := s.DB.WithContext(ctx).Find(&games).Error; err != nil {
		poollog.Event(zap.ErrorLevel, "poolgroup.Store.QueueReset", "s.DB.WithContext.Find.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return err
	}
	symbols := make([]string, 0, len(games))
	for _, g := range games {
		symbols = append(symbols, g.ConfName)
	}
	for _, a := range agents {
		payload, _ := json.Marshal(Plan{Intent: Intent{AgentID: a.Id, Kind: "reset", Symbols: symbols, RequestID: "reset:" + batchID}})
		op := manager.PoolGroupOperation{AgentID: a.Id, RequestID: "reset:" + batchID, Kind: "reset", Payload: string(payload), Status: Pending}
		if err := s.DB.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&op).Error; err != nil {
			poollog.Event(zap.ErrorLevel, "poolgroup.Store.QueueReset", "s.DB.WithContext.Clauses.Create.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return err
		}
	}
	poollog.Event(zap.InfoLevel, "poolgroup.Store.QueueReset", "batch_queued", "重置批次持久化完成，尚需执行清池", append(poolLogFields, zap.Int("agentCount", len(agents)), zap.Int("gameCount", len(symbols)))...)
	return nil
}
