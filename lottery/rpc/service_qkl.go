package rpc

import (
	"app/config"
	"app/entity"
	"context"
	"fmt"
	"lottery/dao"
	"micro_service/services"

	jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func currencyFromCent(v int64) decimal.Decimal {
	return decimal.NewFromInt(v).Div(decimal.NewFromInt(100))
}

func (d *LotteryService) rollbackPoolChange(agentId, userId uint32, symbol, currencyType, recordId string, bet, award, rate decimal.Decimal) {
	dao.CacheIns().ChangePool(int64(agentId), int32(userId), symbol, currencyType, recordId, bet.Neg(), award.Neg(), rate)
}

func (d *LotteryService) qklBet(agentId, userId uint32, exchange decimal.Decimal, symbol, recordId, betStr, currencyType string) (decimal.Decimal, bool) {
	var newCurrency int64 = 0
	bet, _ := decimal.NewFromString(betStr)
	exBet := bet.Mul(exchange)
	pc := config.CfgIns.GetPoolCfg(int64(agentId), symbol)
	if pc == nil {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", recordId), zap.Any("pc", pc))
		return decimal.Zero, false
	}
	zap.L().Debug("qklBet:开始下注", zap.Any("agentId", agentId),
		zap.Any("symbol", symbol),
		zap.Any("roundId", recordId),
		zap.Any("playerId", userId),
		zap.Any("bet", bet),
		zap.Any("currenType", currencyType))
	if bet.GreaterThan(decimal.Zero) {
		//首先扣减用户金额
		tmp, err := d.updatePlayerCurrency(userId, (bet.Neg()).Mul(decimal.NewFromInt(100)).IntPart())
		if err != services.ErrorCode_OK {
			zap.L().Debug("qklBet:下注失败,更新玩家积分失败",
				zap.Any("agentId", agentId),
				zap.Any("symbol", symbol),
				zap.Any("roundId", recordId),
				zap.Any("playerId", userId),
				zap.Any("bet", bet),
				zap.Any("currenType", currencyType))
			return decimal.Zero, false
		}
		newCurrency = tmp
	}
	//
	nc := currencyFromCent(newCurrency)
	dao.CacheIns().ChangePool(int64(agentId), int32(userId), symbol, currencyType, recordId, exBet, decimal.Zero, pc.Pool[1].Revenue)
	user := dao.CacheIns().GetUser(int64(agentId), int64(userId))
	if user != nil && user.IsTourist == 0 {
		if bet.GreaterThan(decimal.Zero) {
			//下注流水
			d.SaveBill(uint32(agentId), userId, bet.Neg(), nc.Truncate(2).InexactFloat64(), symbol, "下注", currencyType, recordId)
		}
		//打点水池记录
		d.pcr.Record(int64(agentId), symbol, dao.CacheIns().GetPool(int64(agentId), symbol))
	}

	zap.L().Debug("qklBet:下注成功",
		zap.Any("agentId", agentId),
		zap.Any("symbol", symbol),
		zap.Any("roundId", recordId),
		zap.Any("playerId", userId))
	return nc, true
}

func (d *LotteryService) qklReturn(agentId, userId uint32, exchange decimal.Decimal, symbol, recordId, betStr, currencyType string) (decimal.Decimal, bool) {
	var newCurrency int64 = 0
	bet, _ := decimal.NewFromString(betStr)
	exBet := bet.Mul(exchange)
	pc := config.CfgIns.GetPoolCfg(int64(agentId), symbol)
	if pc == nil {
		zap.L().Error("获取Pool配置文件失败", zap.Any("roundId", recordId), zap.Any("pc", pc))
		return decimal.Zero, false
	}
	zap.L().Debug("qklReturn:回退", zap.Any("agentId", agentId),
		zap.Any("symbol", symbol),
		zap.Any("roundId", recordId),
		zap.Any("playerId", userId),
		zap.Any("bet", bet),
		zap.Any("currenType", currencyType))
	if bet.GreaterThan(decimal.Zero) {
		//首先扣减用户金额
		tmp, err := d.updatePlayerCurrency(userId, bet.Mul(decimal.NewFromInt(100)).IntPart())
		if err != services.ErrorCode_OK {
			zap.L().Debug("qklReturn:回退失败,更新玩家积分失败",
				zap.Any("agentId", agentId),
				zap.Any("symbol", symbol),
				zap.Any("roundId", recordId),
				zap.Any("playerId", userId),
				zap.Any("bet", bet),
				zap.Any("currenType", currencyType))
			return decimal.Zero, false
		}
		newCurrency = tmp
	}
	//
	nc := currencyFromCent(newCurrency)
	dao.CacheIns().ChangePool(int64(agentId), int32(userId), symbol, currencyType, recordId, exBet.Abs().Neg(), decimal.Zero, pc.Pool[1].Revenue)
	user := dao.CacheIns().GetUser(int64(agentId), int64(userId))
	if user != nil && user.IsTourist == 0 {
		if bet.GreaterThan(decimal.Zero) {
			//下注流水
			d.SaveBill(uint32(agentId), userId, bet.Neg(), nc.Truncate(2).InexactFloat64(), symbol, "回退", currencyType, recordId)
		}
		//打点水池记录
		d.pcr.Record(int64(agentId), symbol, dao.CacheIns().GetPool(int64(agentId), symbol))
	}

	zap.L().Debug("qklReturn:回退成功",
		zap.Any("agentId", agentId),
		zap.Any("symbol", symbol),
		zap.Any("roundId", recordId),
		zap.Any("playerId", userId))
	return nc, true
}

func (d *LotteryService) PoolAmountResult(_ context.Context, req *services.PoolAmountResultReq) (resp *services.PoolAmountResultResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.PoolAmountResultResp{Code: services.ErrorCode_OK}
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("PoolAmountResult:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("PoolAmountResult:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("PoolAmountResult:获取Pool配置文件失败", zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("PoolAmountResult:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("agentId", req.AgentId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	p := dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName)
	//换算成对应币种的积分   cny->[currencyType]
	resp.Currency = p.Div(exchange).Truncate(2).String()
	return resp, nil
}

func (d *LotteryService) QKLDoBetInit(_ context.Context, req *services.QKLDoBetInitReq) (resp *services.QKLDoBetInitResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLDoBetInitResp{Code: services.ErrorCode_OK}
	bet, _ := decimal.NewFromString(req.Bet)
	resp.Code = services.ErrorCode_OK
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("QKLDoBetInit:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("QKLDoBetInit:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLDoBetInit", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLDoBetInit:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", req.RoundID),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	//下注或购买
	if bet.GreaterThan(decimal.Zero) {
		newCurrency, ok := d.qklBet(req.AgentId, req.UserId, exchange, eGame.ConfName, req.RoundID, req.Bet, req.CurrencyType)
		if !ok {
			resp.Code = services.ErrorCode_SYSTEM_ERROR
		} else {
			resp.Currency = newCurrency.String()
		}
	}
	return resp, nil
}

func (d *LotteryService) QKLDoBetMore(_ context.Context, req *services.QKLDoBetMoreReq) (resp *services.QKLDoBetMoreResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLDoBetMoreResp{Code: services.ErrorCode_OK}
	bet, _ := decimal.NewFromString(req.Bet)
	resp.Code = services.ErrorCode_OK
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("QKLDoBetMore:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("QKLDoBetMore:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLDoBetMore", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLDoBetMore:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", req.RoundID),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	//下注或购买
	if bet.GreaterThan(decimal.Zero) {
		newCurrency, ok := d.qklBet(req.AgentId, req.UserId, exchange, eGame.ConfName, req.RoundID, req.Bet, req.CurrencyType)
		if !ok {
			resp.Code = services.ErrorCode_SYSTEM_ERROR
		} else {
			resp.Currency = newCurrency.String()
		}
	}
	return resp, nil
}

func (d *LotteryService) QKLDoBetContinue(_ context.Context, req *services.QKLDoBetContinueReq) (resp *services.QKLDoBetContinueResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLDoBetContinueResp{Code: services.ErrorCode_OK, CanAfford: false}
	deltaWin, _ := decimal.NewFromString(req.DeltaWin)
	resp.Code = services.ErrorCode_OK
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("QKLDoBetContinue:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Debug("QKLDoBetContinue:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLDoBetContinue", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("QKLDoBetContinue:获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLDoBetContinue:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", req.RoundID),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	//判断是否可以开奖
	if deltaWin.GreaterThan(decimal.Zero) {
		if req.GuaranteedWin {
			//必中 直接预扣 修改水池值
			dao.CacheIns().ChangePool(int64(req.AgentId), int32(req.UserId), eGame.ConfName, req.CurrencyType, req.RoundID, decimal.Zero, deltaWin.Mul(exchange), pc.Pool[1].Revenue)
			resp.CanAfford = true
		} else {
			//判断是否可以开奖
			//判断pool是否足够 足够立马扣除 不继续下注 只检查
			if dao.CacheIns().CheckPoolWithOutBet(int64(req.AgentId), eGame.ConfName, req.RoundID, req.CurrencyType, deltaWin.Mul(exchange), decimal.Zero, decimal.Zero, req.UserId) {
				//不够赔 不可以开
				resp.CanAfford = true
			}
		}
	} else {
		resp.CanAfford = true
	}
	return resp, nil
}

func (d *LotteryService) QKLDoBetSettle(_ context.Context, req *services.QKLDoBetSettleReq) (resp *services.QKLDoBetSettleResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLDoBetSettleResp{Code: services.ErrorCode_OK}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetSettle:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLDoBetSettle", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	//获取注单信息
	ur := &entity.UserRecordInfo{}
	err = jsoniter.UnmarshalFromString(req.Result, ur)
	if err != nil {
		zap.L().Error("QKLDoBetSettle:获取注单信息失败",
			zap.Any("userId", req.UserId),
			zap.Any("symbol", eGame.ConfName),
			zap.Any("agentId", req.AgentId),
			zap.Any("state", req.Result), zap.Any("err", err))
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		return resp, nil
	}
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetSettle:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	account := dao.CacheIns().GetPlayerAccount(int64(req.AgentId), int64(req.UserId))
	if account == "" {
		resp.Code = services.ErrorCode_GAME_FROZEN
		zap.L().Error("QKLDoBetSettle:获取账号信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("QKLDoBetSettle:获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("pc", pc), zap.Any("err", err))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLDoBetSettle:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", req.RoundID),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	win, _ := decimal.NewFromString(req.Win)
	exWin := win.Mul(exchange)
	newCurrency := int64(0)
	//命中 玩家赢了 直接增加玩家余额
	if req.Hit {
		// 注意 这里不需要再扣除水池值了，doBetContinue 已经扣除了
		tmp, err := d.updatePlayerCurrency(req.UserId, win.Mul(decimal.NewFromInt(100)).IntPart())
		if err != services.ErrorCode_OK {
			zap.L().Error("QKLDoBetSettle:更新玩家积分失败",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", req.RoundID),
				zap.Any("playerId", req.UserId),
				zap.Any("Win", req.Win),
				zap.Any("TotalBet", req.TotalBet),
				zap.Any("currenType", req.CurrencyType), zap.Any("err", err))
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			return resp, nil
		}
		newCurrency = tmp
		resp.Currency = currencyFromCent(tmp).String()
	} else {
		tmp, err := d.updatePlayerCurrency(req.UserId, 0)
		if err != services.ErrorCode_OK {
			zap.L().Error("QKLDoBetSettle:更新玩家积分失败",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", req.RoundID),
				zap.Any("playerId", req.UserId),
				zap.Any("Win", req.Win),
				zap.Any("TotalBet", req.TotalBet),
				zap.Any("currenType", req.CurrencyType),
				zap.Any("err", err))
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			return resp, nil
		}
		newCurrency = tmp
		resp.Currency = currencyFromCent(tmp).String()
		zap.L().Debug("QKLDoBetSettle:返还Pool", zap.Any("agentId", req.AgentId), zap.Any("userId", req.UserId), zap.Any("symbol", eGame.ConfName), zap.Any("currencyType", req.CurrencyType), zap.Any("delta", exWin))
		//返还pool值
		dao.CacheIns().ChangePool(int64(req.AgentId), int32(req.UserId), eGame.ConfName, req.CurrencyType, req.RoundID, decimal.Zero, exWin.Neg(), pc.Pool[1].Revenue)
		//归还后置为0
		win = decimal.Zero
	}
	user := dao.CacheIns().GetUser(int64(req.AgentId), int64(req.UserId))
	if user != nil && user.IsTourist == 0 {
		//如果玩家输了 这个值 是表示返还给pool的值 不应该记录在注单里面
		nc := decimal.NewFromInt(newCurrency).Div(decimal.NewFromInt(100))
		//增加结算注单
		record := ConvertRecord(
			uint32(req.AgentId),
			req.UserId,
			req.RoundID,
			req.CurrencyType,
			eGame.ConfName,
			account,
			req.Result,
			nc,
			uint32(eAgent.WebId),
			true,
			ur.BetRecord.TotalBetGold,
			win.InexactFloat64())
		d.SaveRecord(record)
		if req.Hit {
			//下注流水
			d.SaveBill(uint32(req.AgentId), req.UserId, win, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "结算", req.CurrencyType, req.RoundID)
		}
		if req.Complete {
			dao.CacheIns().Complete(int64(req.AgentId), req.UserId, eGame.ConfName, decimal.NewFromFloat(ur.BetRecord.TotalBetGold).Mul(exchange), win.Mul(exchange), decimal.Zero)
		}
		//打点水池记录
		d.pcr.Record(int64(req.AgentId), eGame.ConfName, dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName))
	}
	return resp, nil
}

func (d *LotteryService) QKLDoBetSettleWithCheck(_ context.Context, req *services.QKLDoBetSettleWithCheckReq) (resp *services.QKLDoBetSettleWithCheckResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLDoBetSettleWithCheckResp{Code: services.ErrorCode_OK}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetSettleWithCheck:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLDoBetSettleWithCheck", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	//获取注单信息
	ur := &entity.UserRecordInfo{}
	err = jsoniter.UnmarshalFromString(req.Result, ur)
	if err != nil {
		zap.L().Error("QKLDoBetSettleWithCheck:获取注单信息失败",
			zap.Any("userId", req.UserId),
			zap.Any("symbol", eGame.ConfName),
			zap.Any("agentId", req.AgentId),
			zap.Any("state", req.Result))
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		return resp, nil
	}
	initBet, _ := decimal.NewFromString(req.InitBet)
	w2, _ := decimal.NewFromString(req.Win)
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetSettleWithCheck:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	account := dao.CacheIns().GetPlayerAccount(int64(req.AgentId), int64(req.UserId))
	if account == "" {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetSettleWithCheck:获取账号信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("QKLDoBetSettleWithCheck:获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLDoBetSettleWithCheck:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", req.RoundID),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	if req.Hit == "win" {
		if w2.GreaterThan(decimal.Zero) {
			//判断是否可以开奖
			if !dao.CacheIns().CheckPoolWithChange(int64(req.AgentId), eGame.ConfName, req.RoundID, req.CurrencyType, w2.Mul(exchange), decimal.Zero, decimal.Zero, req.UserId) {
				//不够赔 不可以开
				resp.Code = services.ErrorCode_NO_ENOUGH_POOL_MONEY
				return resp, nil
			}
			tmp, err := d.updatePlayerCurrency(req.UserId, w2.Mul(decimal.NewFromInt(100)).IntPart())
			if err != services.ErrorCode_OK {
				d.rollbackPoolChange(req.AgentId, req.UserId, eGame.ConfName, req.CurrencyType, req.RoundID, decimal.Zero, w2.Mul(exchange), pc.Pool[1].Revenue)
				zap.L().Error("QKLDoBetSettleWithCheck:更新玩家积分失败",
					zap.Any("agentId", req.AgentId),
					zap.Any("symbol", eGame.ConfName),
					zap.Any("roundId", req.RoundID),
					zap.Any("playerId", req.UserId),
					zap.Any("Win", req.Win),
					zap.Any("InitBet", req.InitBet),
					zap.Any("currenType", req.CurrencyType))
				resp.Code = services.ErrorCode_SYSTEM_ERROR
				return resp, nil
			}

			//新余额
			nc := currencyFromCent(tmp)
			resp.CanAfford = true
			resp.Currency = nc.String()
			user := dao.CacheIns().GetUser(int64(req.AgentId), int64(req.UserId))
			if user != nil && user.IsTourist == 0 {
				//增加结算注单
				record := ConvertRecord(
					uint32(req.AgentId),
					req.UserId,
					req.RoundID,
					req.CurrencyType,
					eGame.ConfName,
					account,
					req.Result,
					nc,
					uint32(eAgent.WebId),
					true,
					ur.BetRecord.TotalBetGold,
					w2.InexactFloat64())
				d.SaveRecord(record)
				if w2.Mul(exchange).GreaterThan(decimal.Zero) {
					//下注流水
					d.SaveBill(uint32(req.AgentId), req.UserId, w2, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "结算", req.CurrencyType, req.RoundID)
				}
				//打点水池记录
				d.pcr.Record(int64(req.AgentId), eGame.ConfName, dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName))
			}
		} else {
			tmp, code := d.getPlayerCurrency(req.UserId)
			if code != services.ErrorCode_OK {
				resp.Code = services.ErrorCode_SYSTEM_ERROR
				return resp, nil
			}
			resp.CanAfford = true
			resp.Currency = currencyFromCent(tmp).String()
		}
	}
	if req.Hit == "draw" {
		nc, ok := d.qklReturn(req.AgentId, req.UserId, exchange, eGame.ConfName, req.RoundID, req.InitBet, req.CurrencyType)
		if !ok {
			zap.L().Error("QKLDoBetSettleWithCheck:更新玩家积分失败",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", req.RoundID),
				zap.Any("playerId", req.UserId),
				zap.Any("Win", req.Win),
				zap.Any("InitBet", req.InitBet),
				zap.Any("currenType", req.CurrencyType))
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			return resp, nil
		}
		//新余额
		resp.CanAfford = true
		resp.Currency = nc.String()
		user := dao.CacheIns().GetUser(int64(req.AgentId), int64(req.UserId))
		if user != nil && user.IsTourist == 0 {
			//增加结算注单
			record := ConvertRecord(
				uint32(req.AgentId),
				req.UserId,
				req.RoundID,
				req.CurrencyType,
				eGame.ConfName,
				account,
				req.Result,
				nc,
				uint32(eAgent.WebId),
				true,
				ur.BetRecord.TotalBetGold,
				initBet.InexactFloat64())
			d.SaveRecord(record)
			d.SaveBill(uint32(req.AgentId), req.UserId, w2, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "回退", req.CurrencyType, req.RoundID)
			//打点水池记录
		}
	}

	if req.Hit == "lose" {
		//新余额
		resp.CanAfford = true
		user := dao.CacheIns().GetUser(int64(req.AgentId), int64(req.UserId))
		if user != nil && user.IsTourist == 0 {
			//增加结算注单
			record := ConvertRecord(
				uint32(req.AgentId),
				req.UserId,
				req.RoundID,
				req.CurrencyType,
				eGame.ConfName,
				account,
				req.Result,
				decimal.Zero,
				uint32(eAgent.WebId),
				true,
				ur.BetRecord.TotalBetGold,
				0)
			d.SaveRecord(record)
			//打点水池记录
			d.pcr.Record(int64(req.AgentId), eGame.ConfName, dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName))
		}
	}

	if req.Complete {
		dao.CacheIns().Complete(int64(req.AgentId), req.UserId, eGame.ConfName, initBet.Mul(exchange), w2.Mul(exchange), decimal.Zero)
	}
	return resp, nil
}

func (d *LotteryService) QKLDoBetStop(_ context.Context, req *services.QKLDoBetStopReq) (resp *services.QKLDoBetStopResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLDoBetStopResp{Code: services.ErrorCode_OK}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetStop:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLDoBetStop", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetStop:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	account := dao.CacheIns().GetPlayerAccount(int64(req.AgentId), int64(req.UserId))
	if account == "" {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetStop:获取账号信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("QKLDoBetStop:获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLDoBetStop:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", req.RoundID),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	if nc, ok := d.qklReturn(req.AgentId, req.UserId, exchange, eGame.ConfName, req.RoundID, req.InitBet, req.CurrencyType); ok {
		resp.Currency = nc.String()
		return resp, nil
	}
	resp.Code = services.ErrorCode_SYSTEM_ERROR
	return resp, nil

}

func (d *LotteryService) QKLDoBet(_ context.Context, req *services.QKLDoBetReq) (resp *services.QKLDoBetResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLDoBetResp{Code: services.ErrorCode_OK}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBet:游戏被冻结",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLDoBet", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	bet, _ := decimal.NewFromString(req.Bet)
	win, _ := decimal.NewFromString(req.Win)
	if win.LessThan(decimal.Zero) {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBet:参数异常", zap.Any("req", req))
		return resp, nil
	}
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBet:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	account := dao.CacheIns().GetPlayerAccount(int64(req.AgentId), int64(req.UserId))
	if account == "" {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBet:获取账号信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", req.RoundID),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("QKLDoBet:获取Pool配置文件失败", zap.Any("roundId", req.RoundID), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLDoBet:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", req.RoundID),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	newCurrency := decimal.Zero

	if win.GreaterThan(decimal.Zero) {
		revenue := bet.Mul(exchange).Mul(pc.Pool[1].Revenue).Truncate(4)
		//判断pool是否足够 足够立马扣除
		if !dao.CacheIns().CheckPoolWithChange(int64(req.AgentId), eGame.ConfName, req.RoundID, req.CurrencyType, win.Mul(exchange), bet.Mul(exchange), revenue, req.UserId) {
			//不够赔 不可以开
			resp.Code = services.ErrorCode_NO_ENOUGH_POOL_MONEY
			return resp, nil
		}
		var tmp int64 = 0
		var code services.ErrorCode = services.ErrorCode_OK
		tmp, code = d.updatePlayerCurrency(req.UserId, bet.Neg().Mul(decimal.NewFromInt(100)).IntPart())
		if code != services.ErrorCode_OK {
			d.rollbackPoolChange(req.AgentId, req.UserId, eGame.ConfName, req.CurrencyType, req.RoundID, bet.Mul(exchange), win.Mul(exchange), pc.Pool[1].Revenue)
			zap.L().Error("QKLDoBet:更新玩家积分失败",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", req.RoundID),
				zap.Any("playerId", req.UserId),
				zap.Any("bet", bet),
				zap.Any("currenType", req.CurrencyType))
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			return resp, nil
		}
		if bet.GreaterThan(decimal.Zero) {
			nc := currencyFromCent(tmp)
			d.SaveBill(uint32(req.AgentId), req.UserId, bet.Neg(), nc.Truncate(2).InexactFloat64(), eGame.ConfName, "下注", req.CurrencyType, req.RoundID)
		}

		tmp, code = d.updatePlayerCurrency(req.UserId, win.Mul(decimal.NewFromInt(100)).IntPart())
		if code != services.ErrorCode_OK {
			d.rollbackPoolChange(req.AgentId, req.UserId, eGame.ConfName, req.CurrencyType, req.RoundID, bet.Mul(exchange), win.Mul(exchange), pc.Pool[1].Revenue)
			zap.L().Error("QKLDoBet:更新玩家积分失败",
				zap.Any("agentId", req.AgentId),
				zap.Any("symbol", eGame.ConfName),
				zap.Any("roundId", req.RoundID),
				zap.Any("playerId", req.UserId),
				zap.Any("win", win),
				zap.Any("currenType", req.CurrencyType))
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			return resp, nil
		}

		//新余额
		nc := currencyFromCent(tmp)
		if win.GreaterThan(decimal.Zero) {
			d.SaveBill(uint32(req.AgentId), req.UserId, win, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "结算", req.CurrencyType, req.RoundID)
		}
		resp.Currency = nc.String()
		newCurrency = nc
	} else {
		if bet.GreaterThan(decimal.Zero) {
			if nc, ok := d.qklBet(req.AgentId, req.UserId, exchange, eGame.ConfName, req.RoundID, req.Bet, req.CurrencyType); ok {
				resp.Currency = nc.Truncate(2).String()
				newCurrency = nc
			} else {
				resp.Code = services.ErrorCode_SYSTEM_ERROR
			}
		}
	}

	user := dao.CacheIns().GetUser(int64(req.AgentId), int64(req.UserId))
	if user != nil && user.IsTourist == 0 {
		if len(req.Result) > 0 {
			ur := &entity.UserRecordInfo{}
			err = jsoniter.UnmarshalFromString(req.Result, ur)
			if err != nil {
				zap.L().Error("QKLDoBet:获取注单信息失败",
					zap.Any("userId", req.UserId),
					zap.Any("symbol", eGame.ConfName),
					zap.Any("agentId", req.AgentId),
					zap.Any("state", req.Result))
				resp.Code = services.ErrorCode_SYSTEM_ERROR
				return resp, nil
			}
			record := ConvertRecord(
				uint32(req.AgentId),
				req.UserId,
				req.RoundID,
				req.CurrencyType,
				eGame.ConfName,
				account,
				req.Result,
				newCurrency,
				uint32(eAgent.WebId),
				true,
				ur.BetRecord.TotalBetGold,
				win.InexactFloat64())
			d.SaveRecord(record)
		}

		if req.Complete {
			dao.CacheIns().Complete(int64(req.AgentId), req.UserId, eGame.ConfName, bet.Mul(exchange), win.Mul(exchange), decimal.Zero)
		}

		//打点水池记录
		d.pcr.Record(int64(req.AgentId), eGame.ConfName, dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName))
	}
	return resp, nil
}

func (d *LotteryService) QKLDoBetMultiplayerGame(_ context.Context, req *services.QKLDoBetMultiplayerGameReq) (resp *services.QKLDoBetMultiplayerGameResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLDoBetMultiplayerGameResp{Code: services.ErrorCode_OK}
	roundId := fmt.Sprintf("%s#%d", req.RoundID, req.UserId)
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetMultiplayerGame:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLDoBetMultiplayerGame", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	initBet, _ := decimal.NewFromString(req.InitBet)

	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetMultiplayerGame:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	account := dao.CacheIns().GetPlayerAccount(int64(req.AgentId), int64(req.UserId))
	if account == "" {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoBetMultiplayerGame:获取账号信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("QKLDoBetMultiplayerGame:获取Pool配置文件失败", zap.Any("roundId", roundId), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLDoBetMultiplayerGame:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", roundId),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	if initBet.Abs().GreaterThan(decimal.Zero) {
		if nc, ok := d.qklBet(req.AgentId, req.UserId, exchange, eGame.ConfName, roundId, req.InitBet, req.CurrencyType); ok {
			resp.Currency = nc.Truncate(2).String()
		} else {
			zap.L().Error("QKLDoBetMultiplayerGame:获取汇率配置失败",
				zap.Any("currencyType", req.CurrencyType),
				zap.Any("roundId", roundId),
				zap.Any("agentId", req.AgentId),
				zap.Any("playerId", req.UserId),
				zap.Any("gameId", req.GameId))
		}
	}
	return resp, nil
}

func (d *LotteryService) QKLCancelBetMultiplayerGame(_ context.Context, req *services.QKLCancelBetMultiplayerGameReq) (resp *services.QKLCancelBetMultiplayerGameResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLCancelBetMultiplayerGameResp{Code: services.ErrorCode_OK}
	roundId := fmt.Sprintf("%s#%d", req.RoundID, req.UserId)
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLCancelBetMultiplayerGame:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLCancelBetMultiplayerGame", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLCancelBetMultiplayerGame:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	account := dao.CacheIns().GetPlayerAccount(int64(req.AgentId), int64(req.UserId))
	if account == "" {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLCancelBetMultiplayerGame:获取账号信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("QKLCancelBetMultiplayerGame:获取Pool配置文件失败", zap.Any("roundId", roundId), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLCancelBetMultiplayerGame:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", roundId),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}

	if nc, ok := d.qklReturn(req.AgentId, req.UserId, exchange, eGame.ConfName, roundId, req.Bet, req.CurrencyType); ok {
		resp.Currency = nc.Truncate(2).String()
	}
	return resp, nil
}

func (d *LotteryService) QKLDoMultiplayerCashout(_ context.Context, req *services.QKLDoMultiplayerCashoutReq) (resp *services.QKLDoMultiplayerCashoutResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	roundId := fmt.Sprintf("%s#%d", req.RoundID, req.UserId)
	resp = &services.QKLDoMultiplayerCashoutResp{Code: services.ErrorCode_OK}
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eGame == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoMultiplayerCashout:获取游戏信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	zap.L().Debug("QKLDoMultiplayerCashout", zap.Any("symbol", eGame.ConfName), zap.Any("req", req))
	win, _ := decimal.NewFromString(req.Win)
	if win.LessThan(decimal.Zero) {
		win = decimal.Zero
	}

	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	if eAgent == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoMultiplayerCashout:获取代理信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	account := dao.CacheIns().GetPlayerAccount(int64(req.AgentId), int64(req.UserId))
	if account == "" {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		zap.L().Error("QKLDoMultiplayerCashout:获取账号信息失败",
			zap.Any("agentId", req.AgentId),
			zap.Any("roundId", roundId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	pc := config.CfgIns.GetPoolCfg(int64(req.AgentId), eGame.ConfName)
	if pc == nil {
		zap.L().Error("QKLDoMultiplayerCashout:获取Pool配置文件失败", zap.Any("roundId", roundId), zap.Any("pc", pc))
		return resp, nil
	}
	exchange, ok := config.CfgIns.GetExchange(req.CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLDoMultiplayerCashout:获取汇率配置失败",
			zap.Any("currencyType", req.CurrencyType),
			zap.Any("roundId", roundId),
			zap.Any("agentId", req.AgentId),
			zap.Any("playerId", req.UserId),
			zap.Any("gameId", req.GameId))
		return resp, nil
	}
	//判断是否可以开奖
	if !dao.CacheIns().CheckPoolWithChange(int64(req.AgentId), eGame.ConfName, req.RoundID, req.CurrencyType, win.Mul(exchange), decimal.Zero, decimal.Zero, req.UserId) {
		//不够赔 不可以开
		resp.Code = services.ErrorCode_NO_ENOUGH_POOL_MONEY
		return resp, nil
	}
	tmp, code := d.updatePlayerCurrency(req.UserId, win.Mul(decimal.NewFromInt(100)).IntPart())
	if code != services.ErrorCode_OK {
		d.rollbackPoolChange(req.AgentId, req.UserId, eGame.ConfName, req.CurrencyType, req.RoundID, decimal.Zero, win.Mul(exchange), pc.Pool[1].Revenue)
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		return resp, nil
	}
	resp.Currency = currencyFromCent(tmp).String()
	zap.L().Debug("QKLDoMultiplayerCashout:扣除Pool", zap.Any("agentId", req.AgentId), zap.Any("userId", req.UserId), zap.Any("symbol", eGame.ConfName), zap.Any("currencyType", req.CurrencyType), zap.Any("delta", win.Mul(exchange)))
	user := dao.CacheIns().GetUser(int64(req.AgentId), int64(req.UserId))
	if user != nil && user.IsTourist == 0 {
		//新余额
		//不记录注单由游戏统一调用记录注单信息
		nc := currencyFromCent(tmp)
		if win.GreaterThan(decimal.Zero) {
			//下注流水
			d.SaveBill(uint32(req.AgentId), req.UserId, win, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "结算", req.CurrencyType, roundId)
		}
		//打点水池记录
		d.pcr.Record(int64(req.AgentId), eGame.ConfName, dao.CacheIns().GetPool(int64(req.AgentId), eGame.ConfName))
	}

	return resp, nil
}

func (d *LotteryService) QKLSaveMultiplayerRecords(_ context.Context, req *services.QKLSaveMultiplayerRecordsReq) (resp *services.QKLSaveMultiplayerRecordsResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	resp = &services.QKLSaveMultiplayerRecordsResp{Code: services.ErrorCode_OK, Currencys: nil}
	if len(req.Records) <= 0 {
		zap.L().Error("QKLSaveMultiplayerRecords:批量结算", zap.Any("record count", len(req.Records)))
		return &services.QKLSaveMultiplayerRecordsResp{Code: services.ErrorCode_OK, Currencys: nil}, nil
	}
	tGame := dao.GamesManagerIns().GetById(int64(req.Records[0].GameId))
	if tGame == nil {
		return &services.QKLSaveMultiplayerRecordsResp{Code: services.ErrorCode_PARAMS_INVALID, Currencys: nil}, nil
	}
	zap.L().Debug("QKLSaveMultiplayerRecords:批量保存注单数据结算", zap.Any("symbol", tGame.ConfName), zap.Any("record count", len(req.Records)), zap.Any("recordId", req.Records[0].RoundID), zap.Any("req", req))
	ids, tmp := make([]uint32, 0, 64), make(map[uint32]int64)
	for _, item := range req.Records {
		ids = append(ids, item.UserId)
		if len(ids) >= 100 {
			if ncs, e := dao.RedisIns().BatchGetPlayerCurrencys(ids); e == nil {
				for id, c := range ncs {
					tmp[id] = c
				}
			}
			ids = make([]uint32, 0, 64)
		}
	}

	if len(ids) > 0 {
		if ncs, e := dao.RedisIns().BatchGetPlayerCurrencys(ids); e == nil {
			for id, c := range ncs {
				tmp[id] = c
			}
		}
	}

	newCurrencys := make([]*services.QKLNewCurrencyItem, 0)
	for k, v := range tmp {
		newCurrencys = append(newCurrencys, &services.QKLNewCurrencyItem{UserId: k, Currency: decimal.NewFromInt(v).Div(decimal.NewFromInt(100)).Truncate(2).String()})
	}

	for _, item := range req.Records {
		user := dao.CacheIns().GetUser(int64(item.AgentId), int64(item.UserId))
		if user == nil || user.IsTourist != 0 {
			continue
		}
		agent := dao.AgentManagerIns().Get(int64(item.AgentId))
		game := dao.GamesManagerIns().GetById(int64(item.GameId))
		if game.Number > 0 && agent != nil {
			nc := decimal.Zero
			bet, _ := decimal.NewFromString(item.Bet)
			win, _ := decimal.NewFromString(item.Win)
			win = win.Add(bet)
			for _, v := range newCurrencys {
				if v.UserId == item.UserId {
					nc, _ = decimal.NewFromString(v.Currency)
				}
			}
			account := dao.CacheIns().GetPlayerAccount(int64(item.AgentId), int64(item.UserId))
			//增加结算注单
			record := ConvertRecord(
				uint32(item.AgentId),
				item.UserId,
				fmt.Sprintf("%s#%d", item.RoundID, item.UserId),
				item.CurrencyType,
				game.ConfName,
				account,
				item.Log,
				nc,
				uint32(agent.WebId),
				true,
				bet.Truncate(2).InexactFloat64(),
				win.Truncate(2).InexactFloat64())
			d.SaveRecord(record)
		}
	}
	return &services.QKLSaveMultiplayerRecordsResp{Code: services.ErrorCode_OK, Currencys: newCurrencys}, nil
}

func (d *LotteryService) QKLSettleMultiplayer(_ context.Context, req *services.QKLSettleMultiplayerReq) (resp *services.QKLSettleMultiplayerResp, err error) {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error("panic", zap.Any("err", err))
		}
	}()
	if len(req.Records) <= 0 {
		zap.L().Error("QKLSettleMultiplayer:批量结算", zap.Any("record count", len(req.Records)))
		return resp, nil
	}
	newCurrencys := make(map[uint32]*services.QKLNewCurrencyItem)
	deltas := make(map[uint32]int64)
	totalWin, _ := decimal.NewFromString(req.TotalWin)
	exchange, ok := config.CfgIns.GetExchange(req.Records[0].CurrencyType)
	if !ok {
		resp.Code = services.ErrorCode_SYSTEM_ERROR
		zap.L().Error("QKLSettleMultiplayer:获取汇率配置失败",
			zap.Any("currencyType", req.Records[0].CurrencyType),
			zap.Any("agentId", req.Records[0].AgentId),
			zap.Any("playerId", req.Records[0].UserId),
			zap.Any("gameId", req.Records[0].GameId))
		return resp, nil
	}
	totalWin = totalWin.Mul(exchange)
	agentId, gameId := int(req.Records[0].AgentId), int(req.Records[0].GameId)
	//首先判断水池是否足够赔付
	for _, item := range req.Records {
		user := dao.CacheIns().GetUser(int64(item.AgentId), int64(item.UserId))
		if user == nil || user.IsTourist != 0 {
			continue
		}
		win, _ := decimal.NewFromString(item.Win)
		if win.LessThanOrEqual(decimal.Zero) {
			continue
		}
	}

	// 百人类的 可以这么写 没有并发问题
	game := dao.GamesManagerIns().GetById(int64(gameId))
	if game == nil {
		resp.Code = services.ErrorCode_PARAMS_INVALID
		return resp, nil
	}

	zap.L().Debug("QKLSettleMultiplayer:批量结算", zap.Any("symbol", game.ConfName), zap.Any("record count", len(req.Records)), zap.Any("req", req))

	pool := dao.CacheIns().GetPool(int64(agentId), game.ConfName)
	if pool.LessThan(totalWin) {
		zap.L().Debug("QKLSettleMultiplayer:赔付失败", zap.Any("req", req))
		resp.Code = services.ErrorCode_NO_ENOUGH_POOL_MONEY
		return resp, nil
	}
	//批量更新积分
	for _, item := range req.Records {
		roundId := fmt.Sprintf("%s#%d", item.RoundID, item.UserId)
		win, _ := decimal.NewFromString(item.Win)
		game := dao.GamesManagerIns().GetById(int64(item.GameId))
		bet, _ := decimal.NewFromString(item.Bet)
		if win.GreaterThan(decimal.Zero) {
			win = win.Add(bet)
		} else {
			win = decimal.Zero
		}
		//换算成redis里面的单位 这里不能累加 一个玩家 同一局游戏 只能产生一条游戏记录 如果出现多个游戏记录就是bug  不能累加
		deltas[item.UserId] = win.Mul(decimal.NewFromInt(100)).Truncate(0).IntPart()
		if len(deltas) >= 100 {
			tmp, err := dao.RedisIns().BatchUpdatePlayerCurrencys(deltas)
			if err != nil {
				zap.L().Error("QKLSettleMultiplayer:更新玩家积分失败",
					zap.Any("agentId", item.AgentId),
					zap.Any("symbol", game.ConfName),
					zap.Any("roundId", roundId),
					zap.Any("playerId", item.UserId),
					zap.Any("win", win),
					zap.Any("currenType", item.CurrencyType))
				resp.Code = services.ErrorCode_SYSTEM_ERROR
				return resp, nil
			}
			for k, v := range tmp {
				newCurrencys[k] = &services.QKLNewCurrencyItem{UserId: k, Currency: decimal.NewFromInt(v).Div(decimal.NewFromInt(100)).Truncate(2).String()}
			}
			deltas = make(map[uint32]int64)
		}
	}
	if len(deltas) > 0 {
		tmp, err := dao.RedisIns().BatchUpdatePlayerCurrencys(deltas)
		if err != nil {
			zap.L().Error("QKLSettleMultiplayer:更新玩家积分失败", zap.Any("data", deltas))
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			return resp, nil
		}
		for k, v := range tmp {
			newCurrencys[k] = &services.QKLNewCurrencyItem{UserId: k, Currency: decimal.NewFromInt(v).Div(decimal.NewFromInt(100)).Truncate(2).String()}
		}
	}
	//批量保存注单信息
	for _, item := range req.Records {
		roundId := fmt.Sprintf("%s#%d", item.RoundID, item.UserId)
		exchange, ok := config.CfgIns.GetExchange(item.CurrencyType)
		if !ok {
			resp.Code = services.ErrorCode_SYSTEM_ERROR
			zap.L().Error("QKLSettleMultiplayer:获取汇率配置失败",
				zap.Any("currencyType", item.CurrencyType),
				zap.Any("roundId", roundId),
				zap.Any("agentId", item.AgentId),
				zap.Any("playerId", item.UserId),
				zap.Any("gameId", item.GameId))
			continue
		}
		agent := dao.AgentManagerIns().Get(int64(item.AgentId))
		game := dao.GamesManagerIns().GetById(int64(item.GameId))
		if game == nil || agent == nil {
			resp.Code = services.ErrorCode_PARAMS_INVALID
			continue
		}
		win, _ := decimal.NewFromString(item.Win)
		account := dao.CacheIns().GetPlayerAccount(int64(item.AgentId), int64(item.UserId))
		bet, _ := decimal.NewFromString(item.Bet)
		if tmp := newCurrencys[item.UserId]; tmp != nil {
			user := dao.CacheIns().GetUser(int64(item.AgentId), int64(item.UserId))
			if user == nil || user.IsTourist != 0 {
				continue
			}
			if win.GreaterThan(decimal.Zero) {
				win = win.Add(bet)
			} else {
				win = decimal.Zero
			}
			//新余额
			nc, _ := decimal.NewFromString(tmp.Currency)
			defRevenue := decimal.NewFromFloat(dao.DefaultRevenue)
			pc := config.CfgIns.GetPoolCfg(int64(item.AgentId), game.ConfName)
			if pc != nil {
				defRevenue = pc.Pool[1].Revenue
			}
			dao.CacheIns().ChangePool(int64(item.AgentId), int32(item.UserId), game.ConfName, item.CurrencyType, item.RoundID, decimal.Zero, win.Mul(exchange), defRevenue)
			dao.CacheIns().Complete(int64(item.AgentId), item.UserId, game.ConfName, bet.Mul(exchange), win.Mul(exchange), defRevenue)
			if win.GreaterThan(decimal.Zero) {
				//下注流水
				d.SaveBill(uint32(item.AgentId), item.UserId, win, nc.Truncate(2).InexactFloat64(), game.ConfName, "结算", item.CurrencyType, roundId)
			}
			newCurrencys[item.UserId] = &services.QKLNewCurrencyItem{UserId: item.UserId, Currency: nc.Truncate(2).String()}
			if game.Number > 0 {
				//增加结算注单
				record := ConvertRecord(
					uint32(item.AgentId),
					item.UserId,
					roundId,
					item.CurrencyType,
					game.ConfName,
					account,
					item.Log,
					nc,
					uint32(agent.WebId),
					true,
					bet.Truncate(2).InexactFloat64(),
					win.Truncate(2).InexactFloat64())
				d.SaveRecord(record)
			}
		}
	}
	resArr := make([]*services.QKLNewCurrencyItem, 0, 64)
	for _, v := range newCurrencys {
		resArr = append(resArr, v)
	}
	resp.Currencys = resArr
	return resp, nil
}
