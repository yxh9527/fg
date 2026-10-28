package poolgroup

import (
	"app/config"
	"app/tables/manager"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	FollowDefault  = "default"
	UseCustom      = "custom"
	NeedsMigration = "migration"
	Superseded     = "superseded"
	Excluded       = "excluded"
	Conflict       = "conflict"
)

type SchemeRequest struct {
	Scope     string       `json:"scope"`
	WebID     int64        `json:"webId"`
	Version   int64        `json:"version"`
	Groups    []Definition `json:"groups"`
	RequestID string       `json:"requestId"`
}

type AgentSchemeRequest struct {
	AgentIDs  []int64          `json:"agentIds"`
	Action    string           `json:"action"`
	Kind      string           `json:"kind"`
	Group     Definition       `json:"group"`
	Expected  map[string]int64 `json:"expected"`
	RequestID string           `json:"requestId"`
}

type AgentPreview struct {
	AgentID    int64        `json:"agentId"`
	Name       string       `json:"name"`
	Generation int64        `json:"generation"`
	Mode       string       `json:"mode"`
	Groups     []Definition `json:"groups"`
	Changes    []string     `json:"changes"`
	Error      string       `json:"error"`
}

func (input SchemeRequest) Validate() error      { return validateSchemeRequest(input) }
func (input AgentSchemeRequest) Validate() error { return validateAgentRequest(input) }

func fingerprint(value interface{}) string {
	raw, _ := json.Marshal(value)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func validRequestID(id string) bool { return len(id) > 0 && len(id) <= 80 }

func lockSchemeRegistry(tx *gorm.DB) error {
	var row manager.PoolSchemeRegistry
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, 1).Error
}

func selectScheme(tx *gorm.DB, webID int64) (*manager.PoolScheme, error) {
	var rows []manager.PoolScheme
	if err := tx.Where("web_id IN ?", []int64{0, webID}).Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	if len(rows) > 1 {
		return nil, errors.New("默认方案站点范围重叠")
	}
	return &rows[0], nil
}

func loadRevision(tx *gorm.DB, id int64) (manager.PoolSchemeRevision, []Definition, error) {
	var row manager.PoolSchemeRevision
	if err := tx.First(&row, id).Error; err != nil {
		return row, nil, err
	}
	groups, err := decodeDefinitions(row.Groups)
	return row, groups, err
}

func (s Store) CurrentScheme(ctx context.Context, webID int64) (map[string]interface{}, error) {
	var all []manager.PoolScheme
	if err := s.DB.WithContext(ctx).Order("web_id ASC").Find(&all).Error; err != nil {
		return nil, err
	}
	var selected manager.PoolScheme
	groups := []Definition{}
	for _, scheme := range all {
		if scheme.WebID == webID {
			selected = scheme
			break
		}
	}
	if selected.ID > 0 {
		_, defs, err := loadRevision(s.DB.WithContext(ctx), selected.RevisionID)
		if err != nil {
			return nil, err
		}
		groups = defs
	}
	return map[string]interface{}{"scheme": selected, "groups": groups, "schemes": all}, nil
}

func validateSchemeRequest(input SchemeRequest) error {
	if input.Scope != "all" && input.Scope != "site" {
		return errors.New("请明确选择全部站点或指定站点")
	}
	if input.Scope == "all" && input.WebID != 0 || input.Scope == "site" && input.WebID <= 0 {
		return errors.New("方案站点范围无效")
	}
	if input.Version < 0 {
		return errors.New("方案版本无效")
	}
	return ValidateDefinitions(input.Groups)
}

func (s Store) PublishScheme(ctx context.Context, input SchemeRequest, operatorID int64) (*manager.PoolSyncBatch, error) {
	if err := validateSchemeRequest(input); err != nil {
		return nil, err
	}
	if !validRequestID(input.RequestID) {
		return nil, errors.New("请求编号无效")
	}
	var batch manager.PoolSyncBatch
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockSchemeRegistry(tx); err != nil {
			return err
		}
		if found, err := findBatch(tx, input.RequestID, fingerprint(input), &batch); found || err != nil {
			return err
		}
		if input.WebID > 0 {
			var n int64
			if err := tx.Model(&manager.Web{}).Where("id = ? AND isDel = 0", input.WebID).Count(&n).Error; err != nil {
				return err
			}
			if n != 1 {
				return errors.New("站点不存在")
			}
		}
		var scopes []manager.PoolScheme
		if err := tx.Find(&scopes).Error; err != nil {
			return err
		}
		var scheme manager.PoolScheme
		for _, row := range scopes {
			if row.WebID == input.WebID {
				scheme = row
			} else if row.WebID == 0 || input.WebID == 0 {
				return errors.New("已有默认方案与本次站点范围重叠，请选择原方案修改")
			}
		}
		if scheme.Version != input.Version {
			return errors.New("默认方案已更新，请刷新后重试")
		}
		previous := []Definition{}
		if scheme.RevisionID > 0 {
			_, defs, err := loadRevision(tx, scheme.RevisionID)
			if err != nil {
				return err
			}
			previous = defs
		}
		if err := validateDefinitionSources(input.Groups, previous); err != nil {
			return err
		}
		if err := validateDefinitionCatalog(tx, input.Groups); err != nil {
			return err
		}
		if scheme.ID == 0 {
			scheme.WebID = input.WebID
			if err := tx.Create(&scheme).Error; err != nil {
				return err
			}
		}
		raw, _ := json.Marshal(input.Groups)
		revision := manager.PoolSchemeRevision{SchemeID: scheme.ID, Version: scheme.Version + 1, Groups: string(raw), OperatorID: operatorID}
		if err := tx.Create(&revision).Error; err != nil {
			return err
		}
		if err := tx.Model(&scheme).Updates(map[string]interface{}{"version": revision.Version, "revision_id": revision.ID}).Error; err != nil {
			return err
		}
		var maxID int64
		query := tx.Model(&manager.Agent{}).Where("id > 0 AND isDel = 0")
		if input.WebID > 0 {
			query = query.Where("webId = ?", input.WebID)
		}
		if err := query.Select("COALESCE(MAX(id),0)").Scan(&maxID).Error; err != nil {
			return err
		}
		batch = manager.PoolSyncBatch{RequestID: input.RequestID, Fingerprint: fingerprint(input), Kind: "publish", WebID: input.WebID, RevisionID: revision.ID, MaxAgentID: maxID, OperatorID: operatorID}
		return tx.Create(&batch).Error
	})
	if err == nil {
		zap.L().Info("默认水池方案已发布", zap.Int64("batchId", batch.ID), zap.Int64("revisionId", batch.RevisionID), zap.Int64("operatorId", operatorID))
	}
	return &batch, err
}

func findBatch(tx *gorm.DB, requestID, hash string, result *manager.PoolSyncBatch) (bool, error) {
	err := tx.Where("request_id = ?", requestID).First(result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if result.Fingerprint != hash {
		return true, errors.New("请求编号已被其他内容使用")
	}
	return true, nil
}

// RegisterAgent must run in the same transaction as agent creation. The registry lock also fences publishing.
func (s Store) RegisterAgent(tx *gorm.DB, agent manager.Agent) error {
	if err := lockSchemeRegistry(tx); err != nil {
		return err
	}
	_, err := s.ensureBinding(tx, agent)
	return err
}

func (s Store) ensureBinding(tx *gorm.DB, agent manager.Agent) (*manager.PoolAgentBinding, error) {
	if agent.Id <= 0 || agent.IsDel != 0 {
		return nil, nil
	}
	scheme, err := selectScheme(tx, agent.WebId)
	if err != nil {
		return nil, err
	}
	var binding manager.PoolAgentBinding
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&binding, "agent_id = ?", agent.Id).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if binding.AgentID > 0 {
		if binding.Mode == UseCustom {
			if binding.TargetRevision > 0 && binding.AppliedRevision != binding.TargetRevision {
				return &binding, s.queueAutomatic(tx, binding)
			}
			return &binding, nil
		}
		if binding.Mode == FollowDefault && scheme == nil {
			binding.Mode = NeedsMigration
			binding.Status = Conflict
			binding.Error = "代理已移出默认方案范围，请保留专属配置或选择新范围的默认方案"
			return &binding, tx.Save(&binding).Error
		}
		if binding.Mode != FollowDefault || scheme == nil {
			return &binding, nil
		}
		if binding.TargetRevision == scheme.RevisionID {
			if binding.AppliedRevision != binding.TargetRevision {
				return &binding, s.queueAutomatic(tx, binding)
			}
			return &binding, nil
		}
		binding.TargetRevision = scheme.RevisionID
		binding.Generation++
		binding.Status = Pending
		binding.Error = ""
		if err = tx.Save(&binding).Error; err != nil {
			return nil, err
		}
		return &binding, s.queueAutomatic(tx, binding)
	}
	if scheme == nil {
		return nil, nil
	}
	_, defs, err := loadRevision(tx, scheme.RevisionID)
	if err != nil {
		return nil, err
	}
	current, err := s.sqlGroups(tx, agent.Id)
	if err != nil {
		return nil, err
	}
	binding = manager.PoolAgentBinding{AgentID: agent.Id, Mode: FollowDefault, TargetRevision: scheme.RevisionID, Generation: 1, Initializing: len(current) == 0, Status: Pending}
	if len(current) > 0 {
		if mapping, match := MatchDefinitions(defs, current); match {
			if err = saveMappings(tx, agent.Id, mapping); err != nil {
				return nil, err
			}
			binding.AppliedRevision = scheme.RevisionID
			binding.AppliedMode = FollowDefault
			binding.Status = Applied
			binding.Initializing = false
		} else {
			binding.Mode = NeedsMigration
			binding.Status = Conflict
			binding.Error = "现有分组与默认方案不同，请保留专属配置或明确恢复默认"
			binding.Initializing = false
		}
	}
	if err = tx.Create(&binding).Error; err != nil {
		return nil, err
	}
	if binding.Status == Pending {
		err = s.queueAutomatic(tx, binding)
	}
	return &binding, err
}

func (s Store) queueAutomatic(tx *gorm.DB, binding manager.PoolAgentBinding) error {
	id := fmt.Sprintf("auto:%d:%d", binding.AgentID, binding.Generation)
	batch := manager.PoolSyncBatch{RequestID: id, Fingerprint: fingerprint(id), Kind: "initialize", RevisionID: binding.TargetRevision, DispatchDone: true}
	if err := tx.Where("request_id = ?", id).FirstOrCreate(&batch).Error; err != nil {
		return err
	}
	item := manager.PoolSyncItem{BatchID: batch.ID, AgentID: binding.AgentID, RevisionID: binding.TargetRevision, Generation: binding.Generation, Status: Pending}
	return tx.Where("batch_id = ? AND agent_id = ?", batch.ID, binding.AgentID).FirstOrCreate(&item).Error
}

// CheckInitialization reads durable state, so Redis loss or service restart cannot reopen an uninitialized game.
func (s Store) CheckInitialization(ctx context.Context, agentID int64, symbol string) error {
	if agentID == 0 {
		return nil
	}
	var binding manager.PoolAgentBinding
	err := s.DB.WithContext(ctx).First(&binding, "agent_id = ?", agentID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := lockSchemeRegistry(tx); err != nil {
				return err
			}
			var agent manager.Agent
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND isDel = 0", agentID).First(&agent).Error; err != nil {
				return err
			}
			b, err := s.ensureBinding(tx, agent)
			if b != nil {
				binding = *b
			}
			return err
		})
	}
	if err != nil {
		return err
	}
	if !binding.Initializing {
		return nil
	}
	_, defs, err := loadRevision(s.DB.WithContext(ctx), binding.TargetRevision)
	if err != nil {
		return err
	}
	for _, g := range defs {
		if contains(g.Symbols, symbol) {
			return errors.New("代理水池配置初始化中，暂不接受新业务请求")
		}
	}
	return nil
}

func validateAgentRequest(input AgentSchemeRequest) error {
	if len(input.AgentIDs) == 0 || len(input.AgentIDs) > 1000 {
		return errors.New("请选择1至1000个正式代理")
	}
	seen := map[int64]bool{}
	for _, id := range input.AgentIDs {
		if id <= 0 || seen[id] {
			return errors.New("代理编号无效或重复，测试代理请使用独立入口")
		}
		seen[id] = true
	}
	if input.Action != "custom" && input.Action != "restore" && input.Action != "keep" {
		return errors.New("代理配置操作无效")
	}
	if input.Action == "custom" {
		if input.Kind != "create" && input.Kind != "update" && input.Kind != "delete" {
			return errors.New("专属分组操作无效")
		}
		if input.Group.Key == "" {
			return errors.New("请选择分组")
		}
		if input.Kind != "delete" {
			if input.Kind == "create" && (len(input.Group.Symbols) == 0 || input.Group.SourceSymbol != input.Group.Symbols[0]) {
				return errors.New("新组配置来源必须是第一个加入的游戏")
			}
			return ValidateDefinitions([]Definition{input.Group})
		}
	}
	return nil
}

func editDefinitions(current []Definition, input AgentSchemeRequest) ([]Definition, error) {
	result := append([]Definition{}, current...)
	index := -1
	for i, g := range result {
		if g.Key == input.Group.Key {
			index = i
		}
	}
	if input.Kind == "create" {
		if index >= 0 {
			return nil, errors.New("分组标识已存在")
		}
		result = append(result, input.Group)
	} else {
		if index < 0 {
			return nil, errors.New("该代理不存在所选分组，请选择其他代理或分组")
		}
		if input.Kind == "delete" {
			result = append(result[:index], result[index+1:]...)
		} else {
			source := result[index].SourceSymbol
			result[index] = input.Group
			result[index].SourceSymbol = source
		}
	}
	return result, ValidateDefinitions(result)
}

func (s Store) agentTarget(tx *gorm.DB, input AgentSchemeRequest, id int64, lock bool) (AgentPreview, error) {
	preview := AgentPreview{AgentID: id, Groups: []Definition{}, Changes: []string{}}
	var agent manager.Agent
	query := tx
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.Select("id,nickName,webId,isDel").Where("id = ? AND isDel = 0", id).First(&agent).Error; err != nil {
		return preview, errors.New("代理不存在")
	}
	preview.Name = agent.NickName
	var binding manager.PoolAgentBinding
	query = tx
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := query.First(&binding, "agent_id = ?", id).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return preview, err
	}
	preview.Generation = binding.Generation
	preview.Mode = binding.Mode
	var pending int64
	if err := tx.Model(&manager.PoolGroupOperation{}).Where("agent_id = ? AND status <> ?", id, Applied).Count(&pending).Error; err != nil {
		return preview, err
	}
	if pending > 0 || binding.Status == Pending || binding.Status == Applying || binding.Status == Waiting || binding.Status == Failed {
		return preview, errors.New("该代理有未完成同步或水池任务，请等待完成或重试后再修改")
	}
	actual, err := s.sqlGroups(tx, id)
	if err != nil {
		return preview, err
	}
	var links []manager.PoolGroupMapping
	if err = tx.Where("agent_id = ?", id).Find(&links).Error; err != nil {
		return preview, err
	}
	current := definitionsFor(actual, links)
	preview.Groups = current
	if input.Action == "restore" {
		scheme, err := selectScheme(tx, agent.WebId)
		if err != nil {
			return preview, err
		}
		if scheme == nil {
			return preview, errors.New("该代理所属站点尚无默认方案")
		}
		_, preview.Groups, err = loadRevision(tx, scheme.RevisionID)
		if err != nil {
			return preview, err
		}
	} else if input.Action == "custom" {
		preview.Groups, err = editDefinitions(current, input)
		if err != nil {
			return preview, err
		}
	}
	if reflect.DeepEqual(current, preview.Groups) {
		preview.Changes = append(preview.Changes, "分组规则不变，仅切换跟随状态")
	} else {
		preview.Changes = append(preview.Changes, fmt.Sprintf("当前%d组，目标%d组；保留可复用组及余额", len(current), len(preview.Groups)))
	}
	if input.Action == "keep" {
		return preview, nil
	}
	if err := validateDefinitionCatalog(tx, preview.Groups); err != nil {
		return preview, err
	}
	for _, g := range preview.Groups {
		for _, sym := range g.Symbols {
			pc := config.CfgIns.GetIndependentPoolCfg(id, sym)
			if pc == nil || pc.Pool[1] == nil {
				return preview, fmt.Errorf("游戏缺少有效独立水池配置：%s", sym)
			}
		}
	}
	return preview, nil
}

func (s Store) PreviewAgents(ctx context.Context, input AgentSchemeRequest) ([]AgentPreview, error) {
	if err := validateAgentRequest(input); err != nil {
		return nil, err
	}
	result := []AgentPreview{}
	for _, id := range input.AgentIDs {
		p, err := s.agentTarget(s.DB.WithContext(ctx), input, id, false)
		if err != nil {
			p.Error = err.Error()
		}
		result = append(result, p)
	}
	return result, nil
}

func (s Store) SubmitAgents(ctx context.Context, input AgentSchemeRequest, operatorID int64) (*manager.PoolSyncBatch, error) {
	if err := validateAgentRequest(input); err != nil {
		return nil, err
	}
	if !validRequestID(input.RequestID) {
		return nil, errors.New("请求编号无效")
	}
	for _, id := range input.AgentIDs {
		if _, ok := input.Expected[strconv.FormatInt(id, 10)]; !ok {
			return nil, errors.New("请先预检查代理配置")
		}
	}
	var batch manager.PoolSyncBatch
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockSchemeRegistry(tx); err != nil {
			return err
		}
		if found, err := findBatch(tx, input.RequestID, fingerprint(input), &batch); found || err != nil {
			return err
		}
		batch = manager.PoolSyncBatch{RequestID: input.RequestID, Fingerprint: fingerprint(input), Kind: input.Action, DispatchDone: true, OperatorID: operatorID}
		if err := tx.Create(&batch).Error; err != nil {
			return err
		}
		for _, id := range input.AgentIDs {
			p, err := s.agentTarget(tx, input, id, true)
			item := manager.PoolSyncItem{BatchID: batch.ID, AgentID: id, Status: Pending}
			if err == nil && p.Generation != input.Expected[strconv.FormatInt(id, 10)] {
				err = errors.New("代理配置版本已变化，请重新预检查")
			}
			if err != nil {
				item.Status = Conflict
				item.Error = err.Error()
				if err = tx.Create(&item).Error; err != nil {
					return err
				}
				continue
			}
			var revision manager.PoolSchemeRevision
			if input.Action == "restore" {
				var agent manager.Agent
				if err = tx.Select("id,webId").First(&agent, id).Error; err != nil {
					return err
				}
				scheme, err := selectScheme(tx, agent.WebId)
				if err != nil {
					return err
				}
				revision.ID = scheme.RevisionID
			} else {
				raw, _ := json.Marshal(p.Groups)
				revision = manager.PoolSchemeRevision{Groups: string(raw), OperatorID: operatorID, Version: p.Generation + 1}
				if err = tx.Create(&revision).Error; err != nil {
					return err
				}
			}
			var binding manager.PoolAgentBinding
			if err = tx.First(&binding, "agent_id = ?", id).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			binding.AgentID = id
			binding.Mode = UseCustom
			if input.Action == "restore" {
				binding.Mode = FollowDefault
			}
			binding.TargetRevision = revision.ID
			binding.Generation++
			binding.Status = Pending
			binding.Error = ""
			if err = tx.Save(&binding).Error; err != nil {
				return err
			}
			item.Generation = binding.Generation
			item.RevisionID = revision.ID
			if err = tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return &batch, err
}

// DispatchSchemeBatch snapshots existing agents by the published upper ID; later agents use registration tasks.
func (s Store) DispatchSchemeBatch(ctx context.Context, id int64) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockSchemeRegistry(tx); err != nil {
			return err
		}
		var batch manager.PoolSyncBatch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&batch, id).Error; err != nil {
			return err
		}
		if batch.DispatchDone {
			return nil
		}
		var agents []manager.Agent
		query := tx.Select("id,nickName,webId,isDel").Where("id > ? AND id <= ? AND isDel = 0", batch.Cursor, batch.MaxAgentID)
		if batch.WebID > 0 {
			query = query.Where("webId = ?", batch.WebID)
		}
		if err := query.Order("id ASC").Limit(100).Find(&agents).Error; err != nil {
			return err
		}
		for _, agent := range agents {
			binding, err := s.ensureBinding(tx, agent)
			if err != nil {
				return err
			}
			item := manager.PoolSyncItem{BatchID: batch.ID, AgentID: agent.Id, RevisionID: batch.RevisionID, Status: Excluded}
			if binding != nil {
				item.Generation = binding.Generation
				if binding.Mode == FollowDefault && binding.TargetRevision == batch.RevisionID {
					item.Status = Pending
				} else if binding.Mode == NeedsMigration {
					item.Status = Conflict
					item.Error = binding.Error
				} else if binding.Mode == FollowDefault {
					item.Status = Superseded
					item.Error = "已发布更新的默认版本"
				} else {
					item.Error = "该代理使用专属配置"
				}
			}
			if err = tx.Create(&item).Error; err != nil {
				return err
			}
			batch.Cursor = agent.Id
		}
		batch.DispatchDone = len(agents) < 100
		return tx.Save(&batch).Error
	})
}

func terminalSync(status string) bool {
	return status == Applied || status == Superseded || status == Excluded || status == Conflict
}

func (s Store) ProcessSyncItem(ctx context.Context, id int64) error {
	var reference manager.PoolSyncItem
	if err := s.DB.WithContext(ctx).First(&reference, id).Error; err != nil {
		return err
	}
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item := reference
		var agent manager.Agent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id,webId,isDel").First(&agent, item.AgentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return tx.Model(&item).Updates(map[string]interface{}{"status": Excluded, "error": "代理已删除"}).Error
			}
			return err
		}
		var binding manager.PoolAgentBinding
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&binding, "agent_id = ?", item.AgentID).Error; err != nil {
			return err
		}
		// Refresh the item after acquiring its binding lock; another worker may have advanced it.
		if err := tx.First(&item, id).Error; err != nil {
			return err
		}
		if terminalSync(item.Status) {
			return nil
		}
		if item.OperationID > 0 {
			var op manager.PoolGroupOperation
			if err := tx.First(&op, item.OperationID).Error; err != nil {
				return err
			}
			item.Status = op.Status
			item.Error = op.Error
			if op.Status == Applied && item.Generation == binding.Generation && item.RevisionID == binding.TargetRevision {
				binding.AppliedRevision = item.RevisionID
				binding.AppliedMode = binding.Mode
				binding.Initializing = false
				binding.Status = Applied
				binding.Error = ""
			}
		} else if agent.IsDel != 0 {
			item.Status = Excluded
			item.Error = "代理已删除"
		} else if item.Generation != binding.Generation || item.RevisionID != binding.TargetRevision {
			item.Status = Superseded
			item.Error = "已有更新的代理配置目标"
		} else {
			revision, defs, err := loadRevision(tx, item.RevisionID)
			if err != nil {
				return err
			}
			if revision.SchemeID > 0 {
				var scheme manager.PoolScheme
				if err = tx.First(&scheme, revision.SchemeID).Error; err != nil {
					return err
				}
				if scheme.WebID > 0 && scheme.WebID != agent.WebId {
					item.Status = Excluded
					item.Error = "代理已移出方案站点范围"
					return tx.Save(&item).Error
				}
			}
			var op *manager.PoolGroupOperation
			var noChange bool
			err = tx.Transaction(func(work *gorm.DB) error {
				var prepareErr error
				op, noChange, prepareErr = s.prepareLayout(work, item, defs, revision.OperatorID)
				return prepareErr
			})
			if err != nil {
				item.Status = Failed
				if errors.Is(err, errSyncWaiting) {
					item.Status = Waiting
				}
				item.Error = err.Error()
			} else if noChange {
				item.Status = Applied
				item.Error = ""
				binding.AppliedRevision = item.RevisionID
				binding.AppliedMode = binding.Mode
				binding.Initializing = false
				binding.Status = Applied
				binding.Error = ""
			} else {
				item.OperationID = op.ID
				item.Status = Pending
				item.Error = ""
			}
		}
		if item.Generation == binding.Generation && item.RevisionID == binding.TargetRevision && item.Status != Applied {
			binding.Status = item.Status
			binding.Error = item.Error
		}
		if err := tx.Save(&binding).Error; err != nil {
			return err
		}
		return tx.Save(&item).Error
	})
}

// RunSchemeWorker is safe on multiple web-api instances; SQL row locks serialize each transition.
func (s Store) RunSchemeWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	var cursor int64
	for {
		passCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		if err := s.schemeWorkerPass(passCtx, &cursor); err != nil {
			zap.L().Error("水池方案同步巡检失败", zap.Error(err))
		}
		cancel()
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s Store) schemeWorkerPass(ctx context.Context, cursor *int64) error {
	var batches []manager.PoolSyncBatch
	if err := s.DB.WithContext(ctx).Where("dispatch_done = ?", false).Order("updated_at ASC,id ASC").Limit(10).Find(&batches).Error; err != nil {
		return err
	}
	for _, b := range batches {
		if err := s.DispatchSchemeBatch(ctx, b.ID); err != nil {
			zap.L().Error("水池方案批次派发失败", zap.Int64("batchId", b.ID), zap.Error(err))
		}
	}
	var agents []manager.Agent
	if err := s.DB.WithContext(ctx).Select("id,webId,isDel").Where("id > ? AND isDel = 0", *cursor).Order("id ASC").Limit(100).Find(&agents).Error; err != nil {
		return err
	}
	for _, agent := range agents {
		if err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := lockSchemeRegistry(tx); err != nil {
				return err
			}
			var locked manager.Agent
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id,webId,isDel").First(&locked, agent.Id).Error; err != nil {
				return err
			}
			_, err := s.ensureBinding(tx, locked)
			return err
		}); err != nil {
			zap.L().Error("代理水池默认方案补偿失败", zap.Int64("agentId", agent.Id), zap.Error(err))
		}
		*cursor = agent.Id
	}
	if len(agents) < 100 {
		*cursor = 0
	}
	var items []manager.PoolSyncItem
	if err := s.DB.WithContext(ctx).Where("status IN ?", []string{Pending, Waiting, Applying, Failed}).Order("updated_at ASC,id ASC").Limit(100).Find(&items).Error; err != nil {
		return err
	}
	for _, item := range items {
		if err := s.ProcessSyncItem(ctx, item.ID); err != nil {
			zap.L().Error("代理水池方案同步失败", zap.Int64("itemId", item.ID), zap.Int64("agentId", item.AgentID), zap.Error(err))
		}
	}
	return nil
}
