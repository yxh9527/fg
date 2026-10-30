package poolgroup

import (
	"app/config"
	"app/tables/manager"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

func independentConfig(agentID int64, symbol string) bool {
	pc := config.CfgIns.GetIndependentPoolCfg(agentID, symbol)
	return pc != nil && pc.Pool[1] != nil
}

func (s Store) PreviewScheme(ctx context.Context, input SchemeRequest, page, size int) (map[string]interface{}, error) {
	if err := validateSchemeRequest(input); err != nil {
		return nil, err
	}
	tx := s.DB.WithContext(ctx)
	var scopes []manager.PoolScheme
	if err := tx.Find(&scopes).Error; err != nil {
		return nil, err
	}
	previous := []Definition{}
	actualVersion := int64(0)
	for _, scope := range scopes {
		if scope.WebID == input.WebID {
			actualVersion = scope.Version
			_, defs, err := loadRevision(tx, scope.RevisionID)
			if err != nil {
				return nil, err
			}
			previous = defs
			if scope.Version != input.Version {
				return nil, errors.New("默认方案已更新，请刷新")
			}
		} else if scope.WebID == 0 || input.WebID == 0 {
			return nil, errors.New("默认方案范围与已保存方案重叠")
		}
	}
	if input.Version != actualVersion {
		return nil, errors.New("默认方案已更新，请刷新")
	}
	if err := validateDefinitionSources(input.Groups, previous); err != nil {
		return nil, err
	}
	if err := validateDefinitionCatalog(tx, input.Groups); err != nil {
		return nil, err
	}
	if input.WebID > 0 {
		var count int64
		if err := tx.Model(&manager.Web{}).Where("id = ? AND isDel = 0", input.WebID).Count(&count).Error; err != nil {
			return nil, err
		}
		if count != 1 {
			return nil, errors.New("站点不存在")
		}
	}
	query := tx.Model(&manager.Agent{}).Where("id > 0 AND isDel = 0")
	if input.WebID > 0 {
		query = query.Where("webId = ?", input.WebID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var agents []manager.Agent
	if err := query.Select("id,nickName,webId,isDel").Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&agents).Error; err != nil {
		return nil, err
	}
	rows := []AgentPreview{}
	for _, agent := range agents {
		row := AgentPreview{AgentID: agent.Id, Name: agent.NickName, Groups: input.Groups, Changes: []string{}}
		var binding manager.PoolAgentBinding
		if err := tx.First(&binding, "agent_id = ?", agent.Id).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		row.Mode = binding.Mode
		row.Generation = binding.Generation
		if binding.Mode == UseCustom {
			row.Changes = append(row.Changes, "保留代理专属配置，不下发默认方案")
			rows = append(rows, row)
			continue
		}
		current, err := s.sqlGroups(tx, agent.Id)
		if err != nil {
			return nil, err
		}
		if binding.Mode == NeedsMigration {
			row.Error = binding.Error
		} else if binding.AgentID == 0 && len(current) > 0 {
			if _, ok := MatchDefinitions(input.Groups, current); !ok {
				row.Error = "现有分组不同，需选择保留专属配置或恢复默认"
			}
		}
		var pending int64
		if err = tx.Model(&manager.PoolGroupOperation{}).Where("agent_id = ? AND status <> ? AND kind <> ?", agent.Id, Applied, "reset").Count(&pending).Error; err != nil {
			return nil, err
		}
		if pending > 0 {
			row.Changes = append(row.Changes, "等待已有水池操作完成后同步")
		}
		for _, g := range input.Groups {
			for _, sym := range g.Symbols {
				pc := independentConfig(agent.Id, sym)
				if !pc {
					row.Error = fmt.Sprintf("游戏缺少有效独立水池配置：%s", sym)
				}
			}
		}
		row.Changes = append(row.Changes, fmt.Sprintf("当前%d组，目标%d组", len(current), len(input.Groups)))
		currentNames := make([]string, 0, len(current))
		for _, group := range current {
			currentNames = append(currentNames, group.Name)
		}
		targetNames := make([]string, 0, len(input.Groups))
		for _, group := range input.Groups {
			targetNames = append(targetNames, group.Name)
		}
		row.Changes = append(row.Changes, fmt.Sprintf("当前组：%s", groupNamesForPreview(currentNames)), fmt.Sprintf("发布目标组：%s", groupNamesForPreview(targetNames)))
		rows = append(rows, row)
	}
	return map[string]interface{}{"rows": rows, "total": total, "page": page, "pageSize": size}, nil
}

func groupNamesForPreview(names []string) string {
	if len(names) == 0 {
		return "无（独立水池）"
	}
	return strings.Join(names, "、")
}

func (s Store) AgentSchemeDetail(ctx context.Context, id int64) (map[string]interface{}, error) {
	tx := s.DB.WithContext(ctx)
	var binding manager.PoolAgentBinding
	if err := tx.First(&binding, "agent_id = ?", id).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	groups, err := s.sqlGroups(tx, id)
	if err != nil {
		return nil, err
	}
	var links []manager.PoolGroupMapping
	if err = tx.Where("agent_id = ?", id).Find(&links).Error; err != nil {
		return nil, err
	}
	target := []Definition{}
	if binding.TargetRevision > 0 {
		_, target, err = loadRevision(tx, binding.TargetRevision)
		if err != nil {
			return nil, err
		}
	}
	return map[string]interface{}{"binding": binding, "groups": definitionsFor(groups, links), "targetGroups": target, "hasDifference": !reflect.DeepEqual(definitionsFor(groups, links), target)}, nil
}

type BatchView struct {
	manager.PoolSyncBatch
	Counts   map[string]int64 `json:"counts"`
	Total    int64            `json:"total"`
	Complete bool             `json:"complete"`
}

func (s Store) batchView(ctx context.Context, batch manager.PoolSyncBatch) (BatchView, error) {
	view := BatchView{PoolSyncBatch: batch, Counts: map[string]int64{}, Complete: batch.DispatchDone}
	var counts []struct {
		Status string
		Total  int64
	}
	if err := s.DB.WithContext(ctx).Model(&manager.PoolSyncItem{}).Select("status,COUNT(*) AS total").Where("batch_id = ?", batch.ID).Group("status").Scan(&counts).Error; err != nil {
		return view, err
	}
	for _, c := range counts {
		view.Counts[c.Status] = c.Total
		view.Total += c.Total
		if c.Status != Applied && c.Status != Excluded && c.Status != Superseded {
			view.Complete = false
		}
	}
	return view, nil
}

func (s Store) SchemeBatches(ctx context.Context, page, size int, includeAutomatic bool) (map[string]interface{}, error) {
	query := s.DB.WithContext(ctx).Model(&manager.PoolSyncBatch{})
	if !includeAutomatic {
		query = query.Where("kind <> ?", "initialize")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var batches []manager.PoolSyncBatch
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&batches).Error; err != nil {
		return nil, err
	}
	rows := []BatchView{}
	for _, batch := range batches {
		view, err := s.batchView(ctx, batch)
		if err != nil {
			return nil, err
		}
		rows = append(rows, view)
	}
	return map[string]interface{}{"rows": rows, "total": total}, nil
}

func (s Store) SchemeBatchDetail(ctx context.Context, id int64, page, size int) (map[string]interface{}, error) {
	var batch manager.PoolSyncBatch
	if err := s.DB.WithContext(ctx).First(&batch, id).Error; err != nil {
		return nil, err
	}
	view, err := s.batchView(ctx, batch)
	if err != nil {
		return nil, err
	}
	query := s.DB.WithContext(ctx).Model(&manager.PoolSyncItem{}).Where("batch_id = ?", id)
	var total int64
	if err = query.Count(&total).Error; err != nil {
		return nil, err
	}
	rows := []manager.PoolSyncItem{}
	if err = query.Order("agent_id ASC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"batch": view, "rows": rows, "total": total}, nil
}
