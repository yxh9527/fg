package v2

import (
	"app/config"
	"app/entity"
	"app/entity/view"
	"app/tables/manager"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	. "web-api-ex/common"
	"web-api-ex/dao"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/olivere/elastic/v7"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 获取游戏列表
func GameList(ctx *gin.Context) {
	games := make([]*manager.Game, 0)
	dao.Mysql().Manager.Find(&games)
	gis := make([]*entity.GameListItem, 0)
	for _, game := range games {
		gi := &entity.GameListItem{
			Id:       int(game.Id),
			Number:   game.Number,
			GameType: game.GameType,
			Name:     game.Name,
			NameZH:   game.NameZH,
			Symbol:   game.ConfName,
		}
		gis = append(gis, gi)
	}
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Msg: "成功", Data: gis})
}

func ApiConfigList(ctx *gin.Context) {
	var total int64 = 0
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("pageSize"))
	if pageSize <= 0 {
		pageSize = 30
	}
	offset := (page - 1) * pageSize

	apiConfig := make([]*manager.ApiConfig, 0)
	dao.Mysql().Manager.Model(manager.ApiConfig{}).Count(&total)
	dao.Mysql().Manager.Order("id desc").Offset(offset).Limit(pageSize).Find(&apiConfig)

	result := &entity.ApiConfigList{Data: apiConfig, Total: total, LastPage: CountPage(total, int64(pageSize)), Page: int64(page), PerPage: int64(pageSize)}
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: result, Msg: "成功"})
}

func GetGameUrl(ctx *gin.Context) {
	data, err := dao.RedisIns().Get("/config/system")
	if err != nil {
		zap.L().Error("获取客户端地址配置失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: nil, Msg: "失败"})
		return
	}
	sys := &config.SystemConfig{}
	err = jsoniter.UnmarshalFromString(data, sys)
	if err != nil {
		zap.L().Error("解析系统配置失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: nil, Msg: "失败"})
		return
	}
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: sys, Msg: "成功"})
}

func UpdateGameUrl(ctx *gin.Context) {
	gameUrl := ctx.Query("gameUrl")
	replay := ctx.Query("replay")
	arr := strings.Split(gameUrl, ",")
	rs := strings.Split(replay, ",")
	data, err := dao.RedisIns().Get("/config/system")
	if err != nil {
		zap.L().Error("获取客户端地址配置失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: nil, Msg: "失败"})
		return
	}
	sys := &config.SystemConfig{}
	err = jsoniter.UnmarshalFromString(data, sys)
	if err != nil {
		zap.L().Error("解析系统配置失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: nil, Msg: "失败"})
		return
	}
	sys.GameUrls = arr
	sys.Replays = rs
	zap.L().Debug("获取系统配置", zap.Any("sys", sys))
	r, _ := jsoniter.MarshalToString(sys)
	dao.RedisIns().Set("/config/system", r, -1)
	dataStr, _ := jsoniter.MarshalToString(map[string]interface{}{
		"key":  "/config/system",
		"data": r,
	})
	str, _ := jsoniter.MarshalToString(map[string]interface{}{
		"event": "config",
		"data":  dataStr,
	})
	dao.RedisIns().Publish("message", str)
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: nil, Msg: "成功"})
}

func ApiConfigAdd(ctx *gin.Context) {
	name := ctx.Query("name")
	client_api_urls := ctx.Query("client_api_urls")
	hall_urls := ctx.Query("hall_urls")
	max_score, _ := strconv.Atoi(ctx.Query("max_score"))
	min_score, _ := strconv.Atoi(ctx.Query("min_score"))
	apiConfig := &manager.ApiConfig{
		Name:          name,
		Min:           min_score,
		Max:           max_score,
		ClientApiUrls: client_api_urls,
		HallUrls:      hall_urls,
	}
	dao.Mysql().Manager.Create(apiConfig)
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: nil, Msg: "成功"})
}

func ApiConfigDel(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Query("id"))
	dao.Mysql().Manager.Model(manager.ApiConfig{}).Delete("id=?", id)
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: nil, Msg: "成功"})
}

func ApiConfigUpdate(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Query("id"))
	name := ctx.Query("name")
	client_api_urls := ctx.Query("client_api_urls")
	hall_urls := ctx.Query("hall_urls")
	max_score, _ := strconv.Atoi(ctx.Query("max_score"))
	min_score, _ := strconv.Atoi(ctx.Query("min_score"))

	apiConfig := &manager.ApiConfig{
		Id:            int64(id),
		Name:          name,
		Min:           min_score,
		Max:           max_score,
		ClientApiUrls: client_api_urls,
		HallUrls:      hall_urls,
	}

	dao.Mysql().Manager.Save(apiConfig)
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: nil, Msg: "成功"})
}

func GameAgentList(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("pageSize"))
	name := ctx.Query("name")
	list := make([]*manager.Game, 0)
	offset := (page - 1) * pageSize
	var db *gorm.DB = dao.Mysql().Manager.Model(manager.Game{}).Debug()
	if len(name) > 0 {
		tag := "%" + name + "%"
		db = db.Where("name like ? or nameZH like ?", tag, tag)
	}
	var count int64 = 0
	db.Count(&count)
	if page > 0 && pageSize > 0 {
		db = db.Offset(offset).Limit(pageSize)
	}
	db.Find(&list)
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: map[string]interface{}{
		"total": count,
		"list":  list,
	}, Msg: "成功"})
}

func GameUrls(ctx *gin.Context) {
	// agentId, _ := strconv.Atoi(ctx.Query("page"))
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: map[string]interface{}{
		"data": config.CfgIns.GetSystemConfig(),
	}, Msg: "成功"})
}

func ClearGameState(ctx *gin.Context) {
	playerId, _ := strconv.Atoi(ctx.Query("playerId"))
	gameId, _ := strconv.Atoi(ctx.Query("gameId"))
	game := &manager.Game{}
	dao.Mysql().Manager.Where("number=?", gameId).Take(game)
	dao.Es().DeleteByQuery("pp_game_states").Query(elastic.NewBoolQuery().Filter(elastic.NewTermQuery("_id", fmt.Sprintf("%d-%s", playerId, game.ConfName)))).WaitForCompletion(false).Do(context.Background())
	zap.L().Debug("清理玩家指定游戏状态成功", zap.Any("playerId", playerId), zap.Any("name", game.NameZH), zap.Any("symbol", game.ConfName))
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: map[string]interface{}{
		"data": config.CfgIns.GetSystemConfig(),
	}, Msg: "成功"})
}

func QueryOrder(ctx *gin.Context) {
	account := ctx.Query("account")
	roundId := ctx.Query("roundId")
	querys := make([]elastic.Query, 0)
	if account != "" {
		//此平台account也是 nickName
		querys = append(querys, elastic.NewMatchQuery("nickName", account))
	}
	if roundId != "" {
		querys = append(querys, elastic.NewMatchPhraseQuery("roundID", roundId))
	}
	boolQuery := elastic.NewBoolQuery().Must(querys...)
	resp, err := dao.Es().Search().Index("pp_gp_settlement").
		FetchSourceContext(elastic.NewFetchSourceContext(true).Exclude("init", "log")).
		Query(boolQuery).
		Pretty(true).
		Sort("playedDate", false).
		Do(context.Background())
	if err != nil {
		zap.L().Error("获取注单信息失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "获取注单信息失败"})
		return
	}
	settlements := make([]*view.Settlement, 0)
	for _, v := range resp.Hits.Hits {
		b, _ := v.Source.MarshalJSON()
		settlement := &view.Settlement{}
		json.Unmarshal(b, settlement)
		settlements = append(settlements, settlement)
	}
	result := &entity.SettlementList{Data: settlements, Total: resp.Hits.TotalHits.Value}
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Data: result, Msg: "成功"})
}

// asc[]: {"totalProfLoss":0,"totalProfLossRate":0,"totalEffect":0,"controlRate":0,"score":0}
func ChangeStatus(ctx *gin.Context) {
	asc := ctx.Query("asc")
	items := make([]*view.AutoSingleCtrl, 0, 32)
	if asc == "" {
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "参数异常", Data: nil})
		return
	}
	if err := jsoniter.UnmarshalFromString(asc, items); err != nil {
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "参数异常", Data: nil})
		return
	}
	err := dao.RedisIns().Set("/config/autoCtrl", asc, -1)
	if err != nil {
		zap.L().Error("保存自动单控配置失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "保存自动单控配置失败", Data: nil})
		return
	}
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Msg: "成功", Data: nil})
}

func SaveSingleCtrlConfig(ctx *gin.Context) {
	asc := ctx.QueryArray("asc[]")
	items := make([]*view.AutoSingleCtrl, 0, 32)
	if len(asc) <= 0 {
		zap.L().Error("参数为空")
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "参数异常", Data: nil})
		return
	}
	for _, v := range asc {
		a := &view.AutoSingleCtrl{}
		if err := jsoniter.UnmarshalFromString(v, a); err != nil {
			zap.L().Error("解析参数失败", zap.Any("err", err))
			ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "参数异常", Data: nil})
			return
		}
		items = append(items, a)
	}
	str, _ := jsoniter.MarshalToString(items)
	err := dao.RedisIns().Set("/config/autoCtrl", str, -1)
	if err != nil {
		zap.L().Error("保存自动单控配置失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "保存自动单控配置失败", Data: nil})
		return
	}

	dataStr, _ := jsoniter.MarshalToString(map[string]interface{}{
		"key":  "/config/autoCtrl",
		"data": str,
	})
	msg, _ := jsoniter.MarshalToString(map[string]interface{}{
		"event": "config",
		"data":  dataStr,
	})
	dao.RedisIns().Publish("message", msg)

	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Msg: "成功", Data: nil})
}

func GetSingleCtrlConfig(ctx *gin.Context) {
	str, err := dao.RedisIns().Get("/config/autoCtrl")
	if err != nil {
		zap.L().Error("获取自动单控配置失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "保存自动单控配置失败", Data: nil})
		return
	}
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Msg: "成功", Data: str})
}

func SaveGameSettingData(ctx *gin.Context) {
	str := ctx.Query("config")
	ac := &config.AwardConfig{
		OddsConfig: make([]*config.AwardOddsConfigItem, 0, 16),
	}
	err := jsoniter.UnmarshalFromString(str, &ac.OddsConfig)
	if err != nil {
		zap.L().Error("配置异常", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "失败", Data: str})
	} else {
		key := "/config/ctrl/default"
		str, _ = jsoniter.MarshalToString(ac)
		err = dao.RedisIns().Set(key, str, -1)
		if err != nil {
			zap.L().Error("保存配置失败", zap.Any("err", err))
			ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "失败", Data: str})
		} else {
			ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Msg: "成功", Data: str})
		}
	}
}

func GetGameSettingData(ctx *gin.Context) {
	key := "/config/ctrl/default"
	str, err := dao.RedisIns().Get(key)
	if err != nil {
		zap.L().Error("获取配置失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "失败", Data: str})
	} else {
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Msg: "成功", Data: str})
	}
}

func SyncAllPool(ctx *gin.Context) {
	str := ctx.Query("config")
	pool := &config.Pool{}
	if err := jsoniter.UnmarshalFromString(str, pool); err != nil {
		zap.L().Error("解析配置失败", zap.Any("err", err))
		ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "失败", Data: nil})
		return
	}
	games := make([]*manager.Game, 0, 64)
	dao.Mysql().Manager.Model(manager.Game{}).Find(&games)
	n := 0
	piple := dao.RedisIns().Client.Pipeline()
	cs := make(map[string]string)
	for _, g := range games {
		key := fmt.Sprintf("/config/pool/%s", g.ConfName)
		pool.Name = g.Name
		pool.NameZH = g.NameZH
		pool.GameId = int64(g.Number)
		pool.Symbol = g.ConfName
		ps, err := jsoniter.MarshalToString(pool)
		if err != nil {
			continue
		}
		cs[key] = ps
		piple.Set(context.Background(), key, ps, -1)
		n++
		if n >= 100 {
			n = 0
			_, err = piple.Exec(context.Background())
			if err != nil {
				zap.L().Error("保存配置失败", zap.Any("err", err))
				ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "失败", Data: nil})
				return
			}
			piple = dao.RedisIns().Client.Pipeline()
		}
	}
	if n > 0 {
		_, err := piple.Exec(context.Background())
		if err != nil {
			zap.L().Error("保存配置失败", zap.Any("err", err))
			ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusInternalServerError, Msg: "失败", Data: nil})
			return
		}
	}
	//广播更新配置
	for key, c := range cs {
		dataStr, _ := jsoniter.MarshalToString(map[string]interface{}{
			"key":  key,
			"data": c,
		})
		msg, _ := jsoniter.MarshalToString(map[string]interface{}{
			"event": "config",
			"data":  dataStr,
		})
		dao.RedisIns().Publish("message", msg)
	}
	ctx.JSON(http.StatusOK, &entity.Response{Code: http.StatusOK, Msg: "成功", Data: nil})
}
