package v2

import (
	"app/poolgroup"
	"app/tables/manager"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func schemePage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "30"))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 30
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func schemeWebID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.DefaultQuery("webId", "0"), 10, 64)
	if err != nil || id < 0 {
		return 0, errors.New("站点范围无效")
	}
	return id, nil
}

func PoolSchemeGet(c *gin.Context) {
	id, err := schemeWebID(c)
	if err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	data, err := groupStore().CurrentScheme(c.Request.Context(), id)
	poolGroupResponse(c, data, err)
}

func PoolSchemeGames(c *gin.Context) {
	var games []manager.Game
	err := groupStore().DB.WithContext(c.Request.Context()).Select("id,name,nameZH,confName").Order("id ASC").Find(&games).Error
	rows := []gin.H{}
	seen := map[string]bool{}
	for _, game := range games {
		if game.ConfName != "" && !seen[game.ConfName] {
			seen[game.ConfName] = true
			rows = append(rows, gin.H{"symbol": game.ConfName, "name": game.Name, "nameZH": game.NameZH})
		}
	}
	poolGroupResponse(c, rows, err)
}

func PoolSchemePublish(c *gin.Context) {
	var input poolgroup.SchemeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		poolGroupResponse(c, nil, errors.New("请求内容无效"))
		return
	}
	if err := input.Validate(); err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	data, err := groupStore().PublishScheme(c.Request.Context(), input, c.GetInt64("id"))
	poolGroupResponse(c, data, err)
}

func PoolSchemePreview(c *gin.Context) {
	var input poolgroup.SchemeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		poolGroupResponse(c, nil, errors.New("请求内容无效"))
		return
	}
	if err := input.Validate(); err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	page, size := schemePage(c)
	data, err := groupStore().PreviewScheme(c.Request.Context(), input, page, size)
	poolGroupResponse(c, data, err)
}

func PoolSchemeAgentsPreview(c *gin.Context) {
	var input poolgroup.AgentSchemeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		poolGroupResponse(c, nil, errors.New("请求内容无效"))
		return
	}
	if err := input.Validate(); err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	data, err := groupStore().PreviewAgents(c.Request.Context(), input)
	poolGroupResponse(c, data, err)
}

func PoolSchemeAgentsSubmit(c *gin.Context) {
	var input poolgroup.AgentSchemeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		poolGroupResponse(c, nil, errors.New("请求内容无效"))
		return
	}
	if err := input.Validate(); err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	data, err := groupStore().SubmitAgents(c.Request.Context(), input, c.GetInt64("id"))
	poolGroupResponse(c, data, err)
}

type schemeAgentView struct {
	AgentID         int64  `json:"agentId"`
	Name            string `json:"name"`
	WebID           int64  `json:"webId"`
	Mode            string `json:"mode"`
	AppliedMode     string `json:"appliedMode"`
	Generation      int64  `json:"generation"`
	TargetRevision  int64  `json:"targetRevision"`
	AppliedRevision int64  `json:"appliedRevision"`
	Initializing    bool   `json:"initializing"`
	Status          string `json:"status"`
	Error           string `json:"error"`
}

func PoolSchemeAgents(c *gin.Context) {
	webID, err := schemeWebID(c)
	if err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	page, size := schemePage(c)
	query := groupStore().DB.WithContext(c.Request.Context()).Model(&manager.Agent{}).Where("gp_agent.id > 0 AND gp_agent.isDel = 0")
	if webID > 0 {
		query = query.Where("gp_agent.webId = ?", webID)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("gp_agent.nickName LIKE ?", "%"+search+"%")
	}
	query = query.Joins("LEFT JOIN pool_agent_bindings AS binding ON binding.agent_id = gp_agent.id")
	if mode := c.Query("mode"); mode != "" {
		query = query.Where("binding.mode = ?", mode)
	}
	var total int64
	if err = query.Count(&total).Error; err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	rows := []schemeAgentView{}
	err = query.Select("gp_agent.id AS agent_id,gp_agent.nickName AS name,gp_agent.webId AS web_id,binding.mode,binding.applied_mode,binding.generation,binding.target_revision,binding.applied_revision,binding.initializing,binding.status,binding.error").Order("gp_agent.id ASC").Offset((page - 1) * size).Limit(size).Scan(&rows).Error
	poolGroupResponse(c, gin.H{"rows": rows, "total": total, "page": page, "pageSize": size}, err)
}

func PoolSchemeAgentDetail(c *gin.Context) {
	id, err := poolGroupAgent(c)
	if err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	if id <= 0 {
		poolGroupResponse(c, nil, errors.New("测试代理请使用独立管理入口"))
		return
	}
	data, err := groupStore().AgentSchemeDetail(c.Request.Context(), id)
	poolGroupResponse(c, data, err)
}

func PoolSchemeBatches(c *gin.Context) {
	page, size := schemePage(c)
	data, err := groupStore().SchemeBatches(c.Request.Context(), page, size, c.Query("includeAutomatic") == "true")
	poolGroupResponse(c, data, err)
}

func PoolSchemeBatchDetail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		poolGroupResponse(c, nil, errors.New("批次编号无效"))
		return
	}
	page, size := schemePage(c)
	data, err := groupStore().SchemeBatchDetail(c.Request.Context(), id, page, size)
	poolGroupResponse(c, data, err)
}

func PoolSchemeRetry(c *gin.Context) {
	var input struct {
		ID int64 `json:"id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || input.ID <= 0 {
		poolGroupResponse(c, nil, errors.New("批次编号无效"))
		return
	}
	db := groupStore().DB.WithContext(c.Request.Context())
	err := db.Transaction(func(tx *gorm.DB) error {
		var batch manager.PoolSyncBatch
		if err := tx.First(&batch, input.ID).Error; err != nil {
			return err
		}
		return tx.Model(&manager.PoolSyncItem{}).Where("batch_id = ? AND status = ?", input.ID, poolgroup.Failed).Updates(map[string]interface{}{"status": poolgroup.Pending, "error": ""}).Error
	})
	poolGroupResponse(c, gin.H{"batchId": input.ID}, err)
}
