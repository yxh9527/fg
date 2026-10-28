package config

import (
	"app/poollog"
	"fmt"
	"go.uber.org/zap"
	"strings"
)

const MaxPoolGroupGames = 10
const PoolGroupSymbolPrefix = "@pool_group:"

// PoolGroupSnapshot is immutable once installed in Configs.
type PoolGroupSnapshot struct {
	ID           int64    `json:"id"`
	AgentID      int64    `json:"agentId"`
	Name         string   `json:"name"`
	SourceSymbol string   `json:"sourceSymbol"`
	Symbols      []string `json:"symbols"`
	Config       *Pool    `json:"config"`
	Version      int64    `json:"version"`
	CreatedAt    int64    `json:"createdAt"`
	UpdatedAt    int64    `json:"updatedAt"`
}

func PoolGroupSymbol(id int64) (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Int64("id", id))
	poolLogTrace := poollog.BeginDebug("config.PoolGroupSymbol", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return fmt.Sprintf("%s%d", PoolGroupSymbolPrefix, id)
}
func ValidatePoolGroup(name string, symbols []string) (poolLogResult0 error) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.Begin("config.ValidatePoolGroup", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult0)
	if strings.TrimSpace(name) == "" || len([]rune(name)) > 100 {
		return fmt.Errorf("组名称不能为空且不能超过100个字符")
	}
	if len(symbols) < 1 || len(symbols) > MaxPoolGroupGames {
		return fmt.Errorf("每组必须选择1至%d个游戏", MaxPoolGroupGames)
	}
	seen := make(map[string]bool, len(symbols))
	for _, symbol := range symbols {
		if symbol == "" || strings.TrimSpace(symbol) != symbol || strings.HasPrefix(symbol, PoolGroupSymbolPrefix) || len(symbol) > 100 {
			return fmt.Errorf("游戏标识无效")
		}
		if seen[symbol] {
			return fmt.Errorf("不能重复选择同一个游戏")
		}
		seen[symbol] = true
	}
	return nil
}
func (cfg *Configs) ReplacePoolGroups(agentID int64, groups []*PoolGroupSnapshot) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(cfg)...)
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogTrace := poollog.BeginDebug("config.Configs.ReplacePoolGroups", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	cfg.Lock.Lock()
	defer cfg.Lock.Unlock()
	if cfg.PoolGroups == nil {
		cfg.PoolGroups = make(map[int64][]*PoolGroupSnapshot)
	}
	cfg.PoolGroups[agentID] = groups
}
func (cfg *Configs) poolGroupLocked(agentID int64, symbol string) (poolLogResult0 *PoolGroupSnapshot) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(cfg)...)
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogFields = append(poolLogFields, zap.String("symbol", symbol))
	poolLogTrace := poollog.BeginDebug("config.Configs.poolGroupLocked", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	for _, group := range cfg.PoolGroups[agentID] {
		if symbol == PoolGroupSymbol(group.ID) {
			return group
		}
		for _, member := range group.Symbols {
			if member == symbol {
				return group
			}
		}
	}
	return nil
}
func (cfg *Configs) GetPoolGroup(agentID int64, symbol string) (poolLogResult0 *PoolGroupSnapshot) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(cfg)...)
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogFields = append(poolLogFields, zap.String("symbol", symbol))
	poolLogTrace := poollog.BeginDebug("config.Configs.GetPoolGroup", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	cfg.Lock.RLock()
	defer cfg.Lock.RUnlock()
	return cfg.poolGroupLocked(agentID, symbol)
}
func (cfg *Configs) PoolStorageSymbol(agentID int64, symbol string) (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(cfg)...)
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogFields = append(poolLogFields, zap.String("symbol", symbol))
	poolLogTrace := poollog.BeginDebug("config.Configs.PoolStorageSymbol", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	if group := cfg.GetPoolGroup(agentID, symbol); group != nil {
		return PoolGroupSymbol(group.ID)
	}
	return symbol
}

// GetIndependentPoolCfg deliberately bypasses grouping, e.g. when a game leaves a group.
func (cfg *Configs) GetIndependentPoolCfg(agentID int64, symbol string) (poolLogResult0 *Pool) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(cfg)...)
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogFields = append(poolLogFields, zap.String("symbol", symbol))
	poolLogTrace := poollog.BeginDebug("config.Configs.GetIndependentPoolCfg", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	cfg.Lock.RLock()
	defer cfg.Lock.RUnlock()
	return cfg.independentPoolLocked(agentID, symbol)
}
func (cfg *Configs) independentPoolLocked(agentID int64, symbol string) (poolLogResult0 *Pool) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(cfg)...)
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogFields = append(poolLogFields, zap.String("symbol", symbol))
	poolLogTrace := poollog.BeginDebug("config.Configs.independentPoolLocked", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	if cfg.Pool == nil {
		return nil
	}
	if p := cfg.Pool.Agent[fmt.Sprintf("%d-%s", agentID, symbol)]; p != nil {
		return p
	}
	if p := cfg.Pool.Agent[fmt.Sprintf("%d_%s", agentID, symbol)]; p != nil {
		return p
	}
	return cfg.Pool.Default[symbol]
}
