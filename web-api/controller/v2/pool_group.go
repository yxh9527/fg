package v2

import (
	"app/config"
	"app/entity"
	"app/poolgroup"
	"app/poollog"
	"app/tables/manager"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"web-api/dao"
)

func groupStore() (poolLogResult0 poolgroup.Store) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("v2.groupStore", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return poolgroup.Store{DB: dao.Mysql().Manager, Redis: dao.RedisIns().Client}
}
func poolGroupResponse(c *gin.Context, data interface{}, err error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogFields = append(poolLogFields, poollog.MetadataFields(data)...)
	poolLogTrace := poollog.Begin("v2.poolGroupResponse", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&err)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "v2.poolGroupResponse", "check", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		c.JSON(http.StatusOK, &entity.Response{Code: http.StatusBadRequest, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Msg: "成功", Data: data})
}
func poolGroupAgent(c *gin.Context) (poolLogResult0 int64, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogTrace := poollog.Begin("v2.poolGroupAgent", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	id, err := strconv.ParseInt(c.Query("agentId"), 10, 64)
	if err != nil || id <= 0 {
		if err != nil {
			poollog.Event(zap.ErrorLevel, "v2.poolGroupAgent", "strconv.ParseInt", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		}
		return 0, errors.New("请先选择代理")
	}
	var agent manager.Agent
	if err = dao.Mysql().Manager.Where("id = ? AND isDel = 0", id).First(&agent).Error; err != nil {
		poollog.Event(zap.ErrorLevel, "v2.poolGroupAgent", "dao.Mysql.Manager.Where.First.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return 0, errors.New("代理不存在")
	}
	if webID := c.Query("webId"); webID != "" && webID != strconv.FormatInt(agent.WebId, 10) {
		return 0, errors.New("代理不属于当前站点")
	}
	return id, nil
}

type poolGroupView struct {
	*config.PoolGroupSnapshot
	Balance string `json:"balance"`
}

func poolGroupBalance(ctx context.Context, agentID int64, group *config.PoolGroupSnapshot) (poolLogResult0 string, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Int64("agentId", agentID))
	poolLogFields = append(poolLogFields, poollog.MetadataFields(group)...)
	poolLogTrace := poollog.Begin("v2.poolGroupBalance", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	scores := make([]decimal.Decimal, 4)
	member := poolgroup.Member(agentID, config.PoolGroupSymbol(group.ID))
	values, err := readPoolCounters.Run(ctx, dao.RedisIns().Client, poolgroup.StatKeys(), member).StringSlice()
	if err != nil {
		poollog.Event(zap.ErrorLevel, "v2.poolGroupBalance", "readPoolCounters.Run.StringSlice", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return "", err
	}
	for i, value := range values {
		scores[i], err = decimal.NewFromString(value)
		if err != nil {
			poollog.Event(zap.ErrorLevel, "v2.poolGroupBalance", "decimal.NewFromString", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return "", err
		}
	}
	return group.Config.Pool[1].Base.Add(scores[0]).Sub(scores[2]).Sub(scores[3]).String(), nil
}

var readPoolCounters = redis.NewScript(`local values={} for i,key in ipairs(KEYS) do values[i]=redis.call('ZSCORE',key,ARGV[1]) or '0' end return values`)

func PoolGroupList(c *gin.Context) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogTrace := poollog.BeginDebug("v2.PoolGroupList", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	agentID, err := poolGroupAgent(c)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "v2.PoolGroupList", "poolGroupAgent", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, err)
		return
	}
	state, err := groupStore().Load(c.Request.Context(), agentID)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "v2.PoolGroupList", "groupStore.Load", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, err)
		return
	}
	rows := make([]poolGroupView, 0, len(state.Groups))
	for _, group := range state.Groups {
		balance, err := poolGroupBalance(c.Request.Context(), agentID, group)
		if err != nil {
			poollog.Event(zap.ErrorLevel, "v2.PoolGroupList", "poolGroupBalance", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			poolGroupResponse(c, nil, err)
			return
		}
		rows = append(rows, poolGroupView{PoolGroupSnapshot: group, Balance: balance})
	}
	var operations []manager.PoolGroupOperation
	if err = dao.Mysql().Manager.Where("agent_id = ?", agentID).Order("id DESC").Limit(30).Find(&operations).Error; err != nil {
		poollog.Event(zap.ErrorLevel, "v2.PoolGroupList", "dao.Mysql.Manager.Where.Order.Limit.Find.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, err)
		return
	}
	poolGroupResponse(c, gin.H{"groups": rows, "operations": operations, "revision": state.Revision}, nil)
}
func PoolGroupDetail(c *gin.Context) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogTrace := poollog.BeginDebug("v2.PoolGroupDetail", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	agentID, err := poolGroupAgent(c)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "v2.PoolGroupDetail", "poolGroupAgent", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, err)
		return
	}
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		if err != nil {
			poollog.Event(zap.ErrorLevel, "v2.PoolGroupDetail", "strconv.ParseInt", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		}
		poolGroupResponse(c, nil, errors.New("水池组编号无效"))
		return
	}
	state, err := groupStore().Load(c.Request.Context(), agentID)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "v2.PoolGroupDetail", "groupStore.Load", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, err)
		return
	}
	for _, g := range state.Groups {
		if g.ID == id {
			balance, err := poolGroupBalance(c.Request.Context(), agentID, g)
			poolGroupResponse(c, poolGroupView{PoolGroupSnapshot: g, Balance: balance}, err)
			return
		}
	}
	poolGroupResponse(c, nil, errors.New("水池组不存在"))
}
func PoolGroupGames(c *gin.Context) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogTrace := poollog.BeginDebug("v2.PoolGroupGames", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	agentID, err := poolGroupAgent(c)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "v2.PoolGroupGames", "poolGroupAgent", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, err)
		return
	}
	editID, _ := strconv.ParseInt(c.Query("groupId"), 10, 64)
	state, err := groupStore().Load(c.Request.Context(), agentID)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "v2.PoolGroupGames", "groupStore.Load", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, err)
		return
	}
	occupied := map[string]bool{}
	for _, g := range state.Groups {
		if g.ID != editID {
			for _, symbol := range g.Symbols {
				occupied[symbol] = true
			}
		}
	}
	var games []manager.Game
	if err = dao.Mysql().Manager.Order("id ASC").Find(&games).Error; err != nil {
		poollog.Event(zap.ErrorLevel, "v2.PoolGroupGames", "dao.Mysql.Manager.Order.Find.Error", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, err)
		return
	}
	result := make([]gin.H, 0, len(games))
	seen := map[string]bool{}
	for _, g := range games {
		if occupied[g.ConfName] || seen[g.ConfName] || g.ConfName == "" {
			continue
		}
		seen[g.ConfName] = true
		result = append(result, gin.H{"id": g.Id, "number": g.Number, "symbol": g.ConfName, "name": g.Name, "nameZH": g.NameZH, "poolConfig": config.CfgIns.GetIndependentPoolCfg(agentID, g.ConfName)})
	}
	poolGroupResponse(c, result, nil)
}
func poolGroupSubmit(c *gin.Context, kind string) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogFields = append(poolLogFields, zap.String("kind", kind))
	poolLogTrace := poollog.Begin("v2.poolGroupSubmit", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(nil)
	var input poolgroup.Intent
	if err := c.ShouldBindJSON(&input); err != nil {
		poollog.Event(zap.ErrorLevel, "v2.poolGroupSubmit", "c.ShouldBindJSON", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, errors.New("请求内容无效"))
		return
	}
	input.Kind = kind
	poollog.Event(zap.InfoLevel, "v2.poolGroupSubmit", "request_validated", "收到水池组配置操作请求", append(poollog.MetadataFields(input), zap.Int64("operatorId", c.GetInt64("id")))...)
	op, err := groupStore().Submit(c.Request.Context(), input, c.GetInt64("id"))
	poolGroupResponse(c, op, err)
}
func PoolGroupCreate(c *gin.Context) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogTrace := poollog.Begin("v2.PoolGroupCreate", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(nil)
	poolGroupSubmit(c, "create")
}
func PoolGroupUpdate(c *gin.Context) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogTrace := poollog.Begin("v2.PoolGroupUpdate", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(nil)
	poolGroupSubmit(c, "update")
}
func PoolGroupDelete(c *gin.Context) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogTrace := poollog.Begin("v2.PoolGroupDelete", zap.InfoLevel, poolLogFields...)
	defer poolLogTrace.Finish(nil)
	poolGroupSubmit(c, "delete")
}
func PoolGroupOperation(c *gin.Context) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("httpMethod", c.Request.Method))
	poolLogFields = append(poolLogFields, zap.String("path", c.Request.URL.Path))
	poolLogFields = append(poolLogFields, zap.Int64("operatorId", c.GetInt64("id")))
	poolLogFields = append(poolLogFields, zap.String("agentId", c.Query("agentId")))
	poolLogFields = append(poolLogFields, zap.String("groupId", c.Query("groupId")))
	poolLogFields = append(poolLogFields, zap.String("id", c.Query("id")))
	poolLogTrace := poollog.BeginDebug("v2.PoolGroupOperation", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	agentID, err := poolGroupAgent(c)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "v2.PoolGroupOperation", "poolGroupAgent", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		poolGroupResponse(c, nil, err)
		return
	}
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		if err != nil {
			poollog.Event(zap.ErrorLevel, "v2.PoolGroupOperation", "strconv.ParseInt", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		}
		poolGroupResponse(c, nil, errors.New("操作编号无效"))
		return
	}
	var op manager.PoolGroupOperation
	err = dao.Mysql().Manager.Where("id = ? AND agent_id = ?", id, agentID).First(&op).Error
	poolGroupResponse(c, op, err)
}
