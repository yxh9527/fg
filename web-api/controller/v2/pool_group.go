package v2

import (
	"app/config"
	"app/entity"
	"app/poolgroup"
	"app/tables/manager"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"net/http"
	"strconv"
	"web-api/dao"
)

func groupStore() poolgroup.Store {
	return poolgroup.Store{DB: dao.Mysql().Manager, Redis: dao.RedisIns().Client}
}
func poolGroupResponse(c *gin.Context, data interface{}, err error) {
	if err != nil {
		c.JSON(http.StatusOK, &entity.Response{Code: http.StatusBadRequest, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Msg: "成功", Data: data})
}
func poolGroupAgent(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Query("agentId"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("请先选择代理")
	}
	var agent manager.Agent
	if err = dao.Mysql().Manager.Where("id = ? AND isDel = 0", id).First(&agent).Error; err != nil {
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

func poolGroupBalance(ctx context.Context, agentID int64, group *config.PoolGroupSnapshot) (string, error) {
	scores := make([]decimal.Decimal, 4)
	member := poolgroup.Member(agentID, config.PoolGroupSymbol(group.ID))
	values, err := readPoolCounters.Run(ctx, dao.RedisIns().Client, poolgroup.StatKeys(), member).StringSlice()
	if err != nil {
		return "", err
	}
	for i, value := range values {
		scores[i], err = decimal.NewFromString(value)
		if err != nil {
			return "", err
		}
	}
	return group.Config.Pool[1].Base.Add(scores[0]).Sub(scores[2]).Sub(scores[3]).String(), nil
}

var readPoolCounters = redis.NewScript(`local values={} for i,key in ipairs(KEYS) do values[i]=redis.call('ZSCORE',key,ARGV[1]) or '0' end return values`)

func PoolGroupList(c *gin.Context) {
	agentID, err := poolGroupAgent(c)
	if err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	state, err := groupStore().Load(c.Request.Context(), agentID)
	if err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	rows := make([]poolGroupView, 0, len(state.Groups))
	for _, group := range state.Groups {
		balance, err := poolGroupBalance(c.Request.Context(), agentID, group)
		if err != nil {
			poolGroupResponse(c, nil, err)
			return
		}
		rows = append(rows, poolGroupView{PoolGroupSnapshot: group, Balance: balance})
	}
	var operations []manager.PoolGroupOperation
	if err = dao.Mysql().Manager.Where("agent_id = ?", agentID).Order("id DESC").Limit(30).Find(&operations).Error; err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	poolGroupResponse(c, gin.H{"groups": rows, "operations": operations, "revision": state.Revision}, nil)
}
func PoolGroupDetail(c *gin.Context) {
	agentID, err := poolGroupAgent(c)
	if err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		poolGroupResponse(c, nil, errors.New("水池组编号无效"))
		return
	}
	state, err := groupStore().Load(c.Request.Context(), agentID)
	if err != nil {
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
	agentID, err := poolGroupAgent(c)
	if err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	editID, _ := strconv.ParseInt(c.Query("groupId"), 10, 64)
	state, err := groupStore().Load(c.Request.Context(), agentID)
	if err != nil {
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
	var input poolgroup.Intent
	if err := c.ShouldBindJSON(&input); err != nil {
		poolGroupResponse(c, nil, errors.New("请求内容无效"))
		return
	}
	input.Kind = kind
	op, err := groupStore().Submit(c.Request.Context(), input, c.GetInt64("id"))
	poolGroupResponse(c, op, err)
}
func PoolGroupCreate(c *gin.Context) { poolGroupSubmit(c, "create") }
func PoolGroupUpdate(c *gin.Context) { poolGroupSubmit(c, "update") }
func PoolGroupDelete(c *gin.Context) { poolGroupSubmit(c, "delete") }
func PoolGroupOperation(c *gin.Context) {
	agentID, err := poolGroupAgent(c)
	if err != nil {
		poolGroupResponse(c, nil, err)
		return
	}
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		poolGroupResponse(c, nil, errors.New("操作编号无效"))
		return
	}
	var op manager.PoolGroupOperation
	err = dao.Mysql().Manager.Where("id = ? AND agent_id = ?", id, agentID).First(&op).Error
	poolGroupResponse(c, op, err)
}
