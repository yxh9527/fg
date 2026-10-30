package poolgroup

import (
	"app/config"
	"app/tables/manager"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Definition describes grouping only. Runtime configuration and balances remain agent-local.
type Definition struct {
	Key          string   `json:"key"`
	Name         string   `json:"name"`
	Symbols      []string `json:"symbols"`
	SourceSymbol string   `json:"sourceSymbol"`
}

var errSyncWaiting = errors.New("等待该代理已有水池任务完成")

func ValidateDefinitions(groups []Definition) error {
	if len(groups) > 500 {
		return errors.New("一个方案最多500个水池组")
	}
	keys, symbols := map[string]bool{}, map[string]bool{}
	for _, g := range groups {
		if g.Key == "" || len(g.Key) > 80 || strings.TrimSpace(g.Key) != g.Key || keys[g.Key] {
			return errors.New("分组标识缺失或重复")
		}
		keys[g.Key] = true
		if err := config.ValidatePoolGroup(g.Name, g.Symbols); err != nil {
			return err
		}
		if err := config.ValidatePoolGroup(g.Name, []string{g.SourceSymbol}); err != nil {
			return errors.New("配置来源游戏无效")
		}
		for _, symbol := range g.Symbols {
			if symbols[symbol] {
				return fmt.Errorf("游戏重复归组：%s", symbol)
			}
			symbols[symbol] = true
		}
	}
	return nil
}

func validateDefinitionSources(groups, previous []Definition) error {
	old := map[string]string{}
	for _, g := range previous {
		old[g.Key] = g.SourceSymbol
	}
	for _, g := range groups {
		if source, exists := old[g.Key]; exists {
			if g.SourceSymbol != source {
				return errors.New("已有方案组的配置来源不能修改")
			}
		} else if g.SourceSymbol != g.Symbols[0] {
			return errors.New("新增方案组的配置来源必须是第一个加入的游戏")
		}
	}
	return nil
}

func validateDefinitionCatalog(tx *gorm.DB, groups []Definition) error {
	wanted := map[string]bool{}
	for _, g := range groups {
		wanted[g.SourceSymbol] = true
		for _, sym := range g.Symbols {
			wanted[sym] = true
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	symbols := make([]string, 0, len(wanted))
	for sym := range wanted {
		symbols = append(symbols, sym)
	}
	var games []manager.Game
	if err := tx.Select("confName").Where("confName IN ?", symbols).Find(&games).Error; err != nil {
		return err
	}
	for _, game := range games {
		delete(wanted, game.ConfName)
	}
	for sym := range wanted {
		return fmt.Errorf("游戏不存在：%s", sym)
	}
	return nil
}

func decodeDefinitions(raw string) ([]Definition, error) {
	var groups []Definition
	if err := json.Unmarshal([]byte(raw), &groups); err != nil {
		return nil, err
	}
	return groups, ValidateDefinitions(groups)
}

func (s Store) sqlGroups(tx *gorm.DB, agentID int64) ([]*config.PoolGroupSnapshot, error) {
	var rows []manager.PoolGroup
	if err := tx.Where("agent_id = ? AND deleted = ?", agentID, false).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	groups := make([]*config.PoolGroupSnapshot, 0, len(rows))
	for _, row := range rows {
		g, err := readSnapshot(tx, row)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func definitionsFor(groups []*config.PoolGroupSnapshot, mappings []manager.PoolGroupMapping) []Definition {
	keys := map[int64]string{}
	for _, m := range mappings {
		keys[m.GroupID] = m.LogicalKey
	}
	result := make([]Definition, 0, len(groups))
	for _, g := range groups {
		key := keys[g.ID]
		if key == "" {
			key = fmt.Sprintf("legacy:%d", g.ID)
		}
		result = append(result, Definition{Key: key, Name: g.Name, Symbols: append([]string{}, g.Symbols...), SourceSymbol: g.SourceSymbol})
	}
	return result
}

// MatchDefinitions adopts only an unambiguous complete match; configuration snapshots are not compared.
func MatchDefinitions(defs []Definition, current []*config.PoolGroupSnapshot) (map[string]int64, bool) {
	if len(defs) != len(current) {
		return nil, false
	}
	used, result := map[int64]bool{}, map[string]int64{}
	for _, d := range defs {
		var id int64
		for _, g := range current {
			if !used[g.ID] && d.Name == g.Name && d.SourceSymbol == g.SourceSymbol && reflect.DeepEqual(d.Symbols, g.Symbols) {
				if id != 0 {
					return nil, false
				}
				id = g.ID
			}
		}
		if id == 0 {
			return nil, false
		}
		used[id] = true
		result[d.Key] = id
	}
	return result, true
}

func saveMappings(tx *gorm.DB, agentID int64, mappings map[string]int64) error {
	if err := tx.Where("agent_id = ?", agentID).Delete(&manager.PoolGroupMapping{}).Error; err != nil {
		return err
	}
	for key, id := range mappings {
		if err := tx.Create(&manager.PoolGroupMapping{AgentID: agentID, LogicalKey: key, GroupID: id}).Error; err != nil {
			return err
		}
	}
	return nil
}

func finalizeLayout(tx *gorm.DB, agentID int64, groups []*config.PoolGroupSnapshot, mappings map[string]int64) error {
	if err := tx.Where("agent_id = ?", agentID).Delete(&manager.PoolGroupMember{}).Error; err != nil {
		return err
	}
	ids := []int64{}
	for _, g := range groups {
		ids = append(ids, g.ID)
	}
	removed := tx.Model(&manager.PoolGroup{}).Where("agent_id = ? AND deleted = ?", agentID, false)
	if len(ids) > 0 {
		removed = removed.Where("id NOT IN ?", ids)
	}
	if err := removed.Updates(map[string]interface{}{"deleted": true, "version": gorm.Expr("version + 1")}).Error; err != nil {
		return err
	}
	for _, g := range groups {
		raw, err := json.Marshal(g.Config)
		if err != nil {
			return err
		}
		if err = tx.Model(&manager.PoolGroup{}).Where("id = ? AND agent_id = ?", g.ID, agentID).Updates(map[string]interface{}{"name": g.Name, "source_symbol": g.SourceSymbol, "config": string(raw), "version": g.Version, "deleted": false}).Error; err != nil {
			return err
		}
		for i, symbol := range g.Symbols {
			if err = tx.Create(&manager.PoolGroupMember{AgentID: agentID, GroupID: g.ID, Symbol: symbol, Position: i}).Error; err != nil {
				return err
			}
		}
	}
	return saveMappings(tx, agentID, mappings)
}

func nextLayoutState(state *State, op manager.PoolGroupOperation, plan Plan) (*State, []string, error) {
	if state.Revision != plan.BaseRevision {
		return nil, nil, errors.New("整套分组的运行版本已变化")
	}
	oldOwners, newOwners := map[string]int64{}, map[string]int64{}
	oldIDs, newIDs := map[int64]bool{}, map[int64]bool{}
	for _, g := range state.Groups {
		oldIDs[g.ID] = true
		for _, sym := range g.Symbols {
			oldOwners[sym] = g.ID
		}
	}
	for _, g := range plan.Layout {
		if g == nil || g.ID <= 0 || g.AgentID != op.AgentID || g.Config == nil || g.Config.Pool[1] == nil || newIDs[g.ID] {
			return nil, nil, errors.New("整套分组运行配置无效")
		}
		if err := config.ValidatePoolGroup(g.Name, g.Symbols); err != nil {
			return nil, nil, err
		}
		newIDs[g.ID] = true
		for _, sym := range g.Symbols {
			if newOwners[sym] != 0 {
				return nil, nil, errors.New("整套分组存在重复游戏")
			}
			newOwners[sym] = g.ID
		}
	}
	clear := map[string]bool{}
	for sym, id := range oldOwners {
		if newOwners[sym] != id {
			clear[sym] = true
		}
	}
	for sym, id := range newOwners {
		if oldOwners[sym] != id {
			clear[sym] = true
		}
	}
	for id := range oldIDs {
		if !newIDs[id] {
			clear[config.PoolGroupSymbol(id)] = true
		}
	}
	for id := range newIDs {
		if !oldIDs[id] {
			clear[config.PoolGroupSymbol(id)] = true
		}
	}
	members := make([]string, 0, len(clear))
	for sym := range clear {
		members = append(members, Member(op.AgentID, sym))
	}
	sort.Strings(members)
	return &State{Revision: state.Revision + 1, LastOperationID: op.ID, Groups: plan.Layout}, members, nil
}

func layoutOperationGate(state *State, op manager.PoolGroupOperation, plan Plan) Gate {
	gate := Gate{OperationID: op.ID}
	old, next := map[int64]*config.PoolGroupSnapshot{}, map[int64]*config.PoolGroupSnapshot{}
	for _, g := range state.Groups {
		old[g.ID] = g
	}
	for _, g := range plan.Layout {
		next[g.ID] = g
	}
	symbols := map[string]bool{}
	for id, g := range old {
		other := next[id]
		if other == nil || g.Version != other.Version || !reflect.DeepEqual(g.Symbols, other.Symbols) {
			for _, sym := range g.Symbols {
				symbols[sym] = true
			}
		}
	}
	for id, g := range next {
		other := old[id]
		if other == nil || g.Version != other.Version || !reflect.DeepEqual(g.Symbols, other.Symbols) {
			for _, sym := range g.Symbols {
				symbols[sym] = true
			}
		}
	}
	for sym := range symbols {
		gate.Symbols = append(gate.Symbols, sym)
	}
	sort.Strings(gate.Symbols)
	return gate
}

// prepareLayout runs in the agent/binding transaction, so task creation and its identity commit together.
func (s Store) prepareLayout(tx *gorm.DB, item manager.PoolSyncItem, defs []Definition, operatorID int64) (*manager.PoolGroupOperation, bool, error) {
	var pending int64
	if err := tx.Model(&manager.PoolGroupOperation{}).Where("agent_id = ? AND status <> ? AND kind <> ?", item.AgentID, Applied, "reset").Count(&pending).Error; err != nil {
		return nil, false, err
	}
	if pending > 0 {
		return nil, false, errSyncWaiting
	}
	current, err := s.sqlGroups(tx, item.AgentID)
	if err != nil {
		return nil, false, err
	}
	transactionStore := s
	transactionStore.DB = tx
	state, err := transactionStore.Load(tx.Statement.Context, item.AgentID)
	if err != nil {
		return nil, false, err
	}
	if !sameRuntimeGroups(current, state.Groups) {
		return nil, false, errors.New("运行分组与数据库不一致，请先恢复运行配置")
	}
	var links []manager.PoolGroupMapping
	if err = tx.Where("agent_id = ?", item.AgentID).Find(&links).Error; err != nil {
		return nil, false, err
	}
	byKey, byID := map[string]int64{}, map[int64]*config.PoolGroupSnapshot{}
	for _, l := range links {
		byKey[l.LogicalKey] = l.GroupID
	}
	for _, g := range current {
		byID[g.ID] = g
		byKey[fmt.Sprintf("legacy:%d", g.ID)] = g.ID
	}
	// Restoring a default can reuse an exact existing group even when the logical key differs.
	used, mapping := map[int64]bool{}, map[string]int64{}
	for _, d := range defs {
		if g := byID[byKey[d.Key]]; g != nil {
			if used[g.ID] {
				return nil, false, errors.New("多个方案组不能映射到同一实际组")
			}
			mapping[d.Key] = g.ID
			used[g.ID] = true
		}
	}
	for _, d := range defs {
		if mapping[d.Key] != 0 {
			continue
		}
		var candidate int64
		for _, g := range current {
			if !used[g.ID] && reflect.DeepEqual(g.Symbols, d.Symbols) {
				if candidate != 0 {
					return nil, false, errors.New("存在无法唯一对应的实际分组，请先处理差异")
				}
				candidate = g.ID
			}
		}
		if candidate > 0 {
			mapping[d.Key] = candidate
			used[candidate] = true
		}
	}
	unchanged := len(current) == len(defs)
	for _, d := range defs {
		old := byID[mapping[d.Key]]
		if old == nil || old.Name != d.Name || !reflect.DeepEqual(old.Symbols, d.Symbols) {
			unchanged = false
		}
	}
	if unchanged {
		return nil, true, saveMappings(tx, item.AgentID, mapping)
	}
	allSymbols := []string{}
	for _, d := range defs {
		allSymbols = append(allSymbols, d.Symbols...)
		allSymbols = append(allSymbols, d.SourceSymbol)
	}
	if len(allSymbols) > 0 {
		var games []manager.Game
		if err = tx.Where("confName IN ?", allSymbols).Find(&games).Error; err != nil {
			return nil, false, err
		}
		found := map[string]bool{}
		for _, g := range games {
			found[g.ConfName] = true
		}
		for _, sym := range allSymbols {
			if !found[sym] {
				return nil, false, fmt.Errorf("游戏不存在：%s", sym)
			}
			pc := config.CfgIns.GetIndependentPoolCfg(item.AgentID, sym)
			if pc == nil || pc.Pool[1] == nil {
				return nil, false, fmt.Errorf("游戏缺少有效独立水池配置：%s", sym)
			}
		}
	}
	target := make([]*config.PoolGroupSnapshot, 0, len(defs))
	changed := len(current) != len(defs)
	for _, d := range defs {
		if old := byID[mapping[d.Key]]; old != nil {
			g := *old
			g.Name = d.Name
			g.Symbols = append([]string{}, d.Symbols...)
			if g.Name != old.Name || !reflect.DeepEqual(g.Symbols, old.Symbols) {
				g.Version++
				g.UpdatedAt = time.Now().Unix()
				changed = true
			}
			target = append(target, &g)
		} else {
			pc := config.CfgIns.GetIndependentPoolCfg(item.AgentID, d.SourceSymbol)
			raw, err := json.Marshal(pc)
			if err != nil {
				return nil, false, err
			}
			var snapshot config.Pool
			if err = json.Unmarshal(raw, &snapshot); err != nil {
				return nil, false, err
			}
			row := manager.PoolGroup{AgentID: item.AgentID, Name: d.Name, SourceSymbol: d.SourceSymbol, Config: string(raw), Deleted: true}
			if err = tx.Create(&row).Error; err != nil {
				return nil, false, err
			}
			mapping[d.Key] = row.ID
			target = append(target, &config.PoolGroupSnapshot{ID: row.ID, AgentID: item.AgentID, Name: d.Name, SourceSymbol: d.SourceSymbol, Symbols: append([]string{}, d.Symbols...), Config: &snapshot, Version: 1, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt})
			changed = true
		}
	}
	if !changed {
		return nil, true, saveMappings(tx, item.AgentID, mapping)
	}
	plan := Plan{Intent: Intent{AgentID: item.AgentID, Kind: "layout", RequestID: fmt.Sprintf("sync:%d", item.ID)}, Layout: target, BaseRevision: state.Revision, Mappings: mapping}
	before, _ := json.Marshal(current)
	raw, err := json.Marshal(plan)
	if err != nil {
		return nil, false, err
	}
	op := manager.PoolGroupOperation{AgentID: item.AgentID, Kind: "layout", RequestID: plan.Intent.RequestID, Payload: string(raw), Before: string(before), Status: Pending, OperatorID: operatorID}
	if err = tx.Create(&op).Error; err != nil {
		return nil, false, err
	}
	return &op, false, nil
}

func sameRuntimeGroups(expected, actual []*config.PoolGroupSnapshot) bool {
	if len(expected) != len(actual) {
		return false
	}
	byID := map[int64]*config.PoolGroupSnapshot{}
	for _, g := range actual {
		if g == nil || byID[g.ID] != nil {
			return false
		}
		byID[g.ID] = g
	}
	for _, g := range expected {
		other := byID[g.ID]
		if other == nil || other.AgentID != g.AgentID || other.Version != g.Version || other.Name != g.Name || other.SourceSymbol != g.SourceSymbol || !reflect.DeepEqual(other.Symbols, g.Symbols) || !reflect.DeepEqual(other.Config, g.Config) {
			return false
		}
	}
	return true
}
