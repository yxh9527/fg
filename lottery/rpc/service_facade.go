package rpc

import (
	"context"
	"strings"

	"app/config"
	"lottery/dao"

	jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"micro_service/services"
)

func parseAmount(v string) decimal.Decimal {
	d, ok := parseAmountStrict(v)
	if !ok {
		return decimal.Zero
	}
	return d
}

// parseAmountStrict 空串视为 0；非法数字返回 false。
func parseAmountStrict(v string) (decimal.Decimal, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return decimal.Zero, true
	}
	d, err := decimal.NewFromString(v)
	if err != nil {
		return decimal.Zero, false
	}
	return d, true
}

func failSlotsDoBet(resp *services.SlotsDoBetResp, code services.ErrorCode, errName string) (*services.SlotsDoBetResp, error) {
	resp.Code = code
	resp.Ret = false
	if errName != "" {
		resp.Error = errName
	} else {
		resp.Error = slotErrorName(code)
	}
	return resp, nil
}

func validateFlowControl(fc *services.SlotDoBetFlowControl) (string, bool) {
	if fc == nil {
		return "", true
	}
	if refund, ok := parseAmountStrict(fc.RefundPoolAmount); !ok {
		return "invalid refundPoolAmount", false
	} else if refund.IsNegative() {
		return "refundPoolAmount can not be negative", false
	}
	for _, key := range fc.DeleteGameStorageKeys {
		if strings.TrimSpace(key) == "" {
			return "deleteGameStorageKeys contains empty key", false
		}
	}
	return "", true
}

func validateGameStorageItems(items []*services.GameStorageItem) (string, bool) {
	for _, item := range items {
		if item == nil {
			return "gameStorage item is nil", false
		}
		if strings.TrimSpace(item.StorageKey) == "" {
			return "gameStorage.storageKey required", false
		}
	}
	return "", true
}

// validateFacadeIdentity 结算门面共用的用户/代理/游戏/币种校验。
// agentId=0 视为测试代理：跳过代理存在/冻结校验，其余字段仍严格校验。
func validateFacadeIdentity(userId, agentId, gameId uint32, roundId, currencyType string) (services.ErrorCode, string) {
	if userId == 0 || gameId == 0 {
		return services.ErrorCode_PARAMS_INVALID, ""
	}
	if strings.TrimSpace(roundId) == "" || strings.TrimSpace(currencyType) == "" {
		return services.ErrorCode_PARAMS_INVALID, ""
	}
	if agentId > 0 && dao.AgentManagerIns().Get(int64(agentId)) == nil {
		return services.ErrorCode_AGENT_FROZEN, ""
	}
	if dao.GamesManagerIns().GetById(int64(gameId)) == nil {
		return services.ErrorCode_PARAMS_INVALID, ""
	}
	if _, ok := config.CfgIns.GetExchange(currencyType); !ok {
		return services.ErrorCode_SYSTEM_ERROR, "exchangeConfigMissing"
	}
	return services.ErrorCode_OK, ""
}

// validateSlotsFacadeBase 普通局/免费局门面共用的基础校验。
func validateSlotsFacadeBase(userId, agentId, gameId uint32, roundId, currencyType string, fc *services.SlotDoBetFlowControl, storages []*services.GameStorageItem) (services.ErrorCode, string) {
	if code, msg := validateFacadeIdentity(userId, agentId, gameId, roundId, currencyType); code != services.ErrorCode_OK {
		return code, msg
	}
	if msg, ok := validateFlowControl(fc); !ok {
		return services.ErrorCode_PARAMS_INVALID, msg
	}
	if msg, ok := validateGameStorageItems(storages); !ok {
		return services.ErrorCode_PARAMS_INVALID, msg
	}
	return services.ErrorCode_OK, ""
}

func failFruitDoBet(resp *services.FruitDoBetResp, code services.ErrorCode) (*services.FruitDoBetResp, error) {
	resp.Code = code
	resp.Ret = false
	return resp, nil
}

func failFruitDoBetMulti(resp *services.FruitDoBetMultiResp, code services.ErrorCode) (*services.FruitDoBetMultiResp, error) {
	resp.Code = code
	resp.Ret = false
	return resp, nil
}

func failFruitRefundMulti(resp *services.FruitRefundMultiResp, code services.ErrorCode) (*services.FruitRefundMultiResp, error) {
	resp.Code = code
	resp.Ret = false
	return resp, nil
}

func failFruitSettleRound(resp *services.FruitSettleRoundResp, code services.ErrorCode) (*services.FruitSettleRoundResp, error) {
	resp.Code = code
	resp.Ret = false
	return resp, nil
}

// validateFacadeRoom 百人房间维度校验（不含单个 userId）。
// agentId=0 视为测试代理：跳过代理存在/冻结校验。
func validateFacadeRoom(agentId, gameId uint32, roundId, currencyType string) (services.ErrorCode, string) {
	if gameId == 0 {
		return services.ErrorCode_PARAMS_INVALID, ""
	}
	if strings.TrimSpace(roundId) == "" || strings.TrimSpace(currencyType) == "" {
		return services.ErrorCode_PARAMS_INVALID, ""
	}
	if agentId > 0 && dao.AgentManagerIns().Get(int64(agentId)) == nil {
		return services.ErrorCode_AGENT_FROZEN, ""
	}
	if dao.GamesManagerIns().GetById(int64(gameId)) == nil {
		return services.ErrorCode_PARAMS_INVALID, ""
	}
	if _, ok := config.CfgIns.GetExchange(currencyType); !ok {
		return services.ErrorCode_SYSTEM_ERROR, "exchangeConfigMissing"
	}
	return services.ErrorCode_OK, ""
}

func (d *LotteryService) applyGameStorageWrites(userId, gameId uint32, currency string, items []*services.GameStorageItem, deleteKeys []string) services.ErrorCode {
	for _, key := range deleteKeys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		if err := d.db.DeleteGameStorage(userId, gameId, currency, key); err != nil {
			zap.L().Error("facade delete game storage failed", zap.Uint32("userId", userId), zap.String("key", key), zap.Error(err))
			return services.ErrorCode_SYSTEM_ERROR
		}
	}
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.StorageKey) == "" {
			continue
		}
		if err := d.db.SaveGameStorage(userId, gameId, currency, item.StorageKey, item.Payload, item.ExpireSeconds); err != nil {
			zap.L().Error("facade save game storage failed", zap.Uint32("userId", userId), zap.String("key", item.StorageKey), zap.Error(err))
			return services.ErrorCode_SYSTEM_ERROR
		}
	}
	return services.ErrorCode_OK
}

func (d *LotteryService) refundFacadePool(agentId, userId, gameId uint32, currencyType, amount string) services.ErrorCode {
	refund := parseAmount(amount)
	if !refund.GreaterThan(decimal.Zero) {
		return services.ErrorCode_OK
	}
	eGame := dao.GamesManagerIns().GetById(int64(gameId))
	if eGame == nil {
		return services.ErrorCode_PARAMS_INVALID
	}
	exchange, ok := config.CfgIns.GetExchange(currencyType)
	if !ok {
		return services.ErrorCode_SYSTEM_ERROR
	}
	dao.CacheIns().ReturnPool(int64(agentId), userId, eGame.ConfName, refund.Mul(exchange))
	d.pcr.Record(int64(agentId), eGame.ConfName, dao.CacheIns().GetPool(int64(agentId), eGame.ConfName))
	return services.ErrorCode_OK
}

func fillBalanceResp(currency string) (string, int64) {
	nc := parseAmount(currency)
	return nc.Truncate(2).StringFixed(2), nc.Mul(decimal.NewFromInt(100)).IntPart()
}

func slotErrorName(code services.ErrorCode) string {
	switch code {
	case services.ErrorCode_NO_ENOUGH_MONEY:
		return "balanceInsufficient"
	case services.ErrorCode_NO_ENOUGH_POOL_MONEY:
		return "poolInsufficient"
	case services.ErrorCode_PARAMS_INVALID:
		return "paramsInvalid"
	default:
		if code != services.ErrorCode_OK {
			return "systemError"
		}
		return ""
	}
}

// SlotsDoBet 新服普通局门面。
// 规则 A：内部走原 SlotsLottery（下注 + Complete 结算/ES 注单），再提交 GameStorage；
// preWinAmount 映射为 MaxProfitLoss；flowControl.refundPoolAmount 在结算成功后按原 ReturnPool 退回。
func (d *LotteryService) SlotsDoBet(ctx context.Context, req *services.SlotsDoBetReq) (resp *services.SlotsDoBetResp, err error) {
	resp = &services.SlotsDoBetResp{Code: services.ErrorCode_OK, Ret: false}
	if req == nil {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "")
	}
	if code, msg := validateSlotsFacadeBase(req.UserId, req.AgentId, req.GameId, req.RoundId, req.CurrencyType, req.FlowControl, req.GameStorage); code != services.ErrorCode_OK {
		return failSlotsDoBet(resp, code, msg)
	}

	bet, betOK := parseAmountStrict(req.BetAmount)
	if !betOK {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "invalid betAmount")
	}
	if bet.IsNegative() {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "betAmount can not be negative")
	}
	win, winOK := parseAmountStrict(req.WinAmount)
	if !winOK {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "invalid winAmount")
	}
	if win.IsNegative() {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "winAmount can not be negative")
	}
	preWin, preOK := parseAmountStrict(req.PreWinAmount)
	if !preOK {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "invalid preWinAmount")
	}
	if preWin.IsNegative() {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "preWinAmount can not be negative")
	}

	idemKey := buildIdempotencyKey("slotsDoBet", u32Str(req.UserId), u32Str(req.GameId), req.CurrencyType, req.RoundId)
	idemSig := idempotencySignature(bet.String(), win.String(), preWin.String(), req.RecordJson, req.Account)
	if hit, payload, code := d.beginIdempotency(idemKey, idemSig); code != services.ErrorCode_OK {
		msg := ""
		if code == services.ErrorCode_PARAMS_INVALID {
			msg = "idempotencyConflict"
		}
		return failSlotsDoBet(resp, code, msg)
	} else if hit {
		if cached, rCode := restoreSlotsDoBetResp(payload); rCode != services.ErrorCode_OK {
			return failSlotsDoBet(resp, rCode, "")
		} else {
			return cached, nil
		}
	}

	slotsResp, callErr := d.SlotsLottery(ctx, &slotsLotteryReq{
		PlayerId:      req.UserId,
		CurrencyType:  req.CurrencyType,
		AgentId:       int64(req.AgentId),
		GameId:        req.GameId,
		ProfitLoss:    win.String(),
		Bet:           bet.String(),
		State:         req.RecordJson,
		RoundID:       req.RoundId,
		MaxProfitLoss: preWin.String(),
		Complete:      true,
		Account:       req.Account,
	})
	if callErr != nil {
		d.abortIdempotency(idemKey)
		zap.L().Error("SlotsDoBet call SlotsLottery failed",
			zap.Uint32("userId", req.UserId),
			zap.String("roundId", req.RoundId),
			zap.Error(callErr))
		return failSlotsDoBet(resp, services.ErrorCode_SYSTEM_ERROR, "")
	}
	if slotsResp == nil {
		d.abortIdempotency(idemKey)
		return failSlotsDoBet(resp, services.ErrorCode_SYSTEM_ERROR, "")
	}
	if !slotsResp.Result || slotsResp.Code != services.ErrorCode_OK {
		d.abortIdempotency(idemKey)
		resp.Code = slotsResp.Code
		resp.Ret = false
		resp.Error = slotErrorName(slotsResp.Code)
		if slotsResp.NewCurrency != "" {
			resp.Currency, resp.CurrencyCent = fillBalanceResp(slotsResp.NewCurrency)
		}
		return resp, nil
	}

	var deleteKeys []string
	if req.FlowControl != nil {
		deleteKeys = req.FlowControl.DeleteGameStorageKeys
		if code := d.refundFacadePool(req.AgentId, req.UserId, req.GameId, req.CurrencyType, req.FlowControl.RefundPoolAmount); code != services.ErrorCode_OK {
			d.abortIdempotency(idemKey)
			return failSlotsDoBet(resp, code, "")
		}
	}
	if code := d.applyGameStorageWrites(req.UserId, req.GameId, req.CurrencyType, req.GameStorage, deleteKeys); code != services.ErrorCode_OK {
		d.abortIdempotency(idemKey)
		return failSlotsDoBet(resp, code, "")
	}

	resp.Ret = true
	resp.Currency, resp.CurrencyCent = fillBalanceResp(slotsResp.NewCurrency)
	if raw, mErr := jsoniter.MarshalToString(resp); mErr == nil {
		d.commitIdempotency(idemKey, idemSig, raw)
	} else {
		d.abortIdempotency(idemKey)
	}
	zap.L().Debug("SlotsDoBet success",
		zap.Uint32("userId", req.UserId),
		zap.Uint32("agentId", req.AgentId),
		zap.Uint32("gameId", req.GameId),
		zap.String("roundId", req.RoundId),
		zap.String("bet", bet.String()),
		zap.String("win", win.String()),
		zap.String("currency", resp.Currency))
	return resp, nil
}

// SlotsDoBetFree 免费/奖励局门面。
// 不重复扣本金。仅当 req.Complete=true 时走原 SlotsLottery.Complete（含统计）；
// Complete=false 但有赢分/注单时只入账+写 ES 注单，不跑 Complete 统计；
// 纯状态 action 可只提交 flowControl/gameStorage。
func (d *LotteryService) SlotsDoBetFree(ctx context.Context, req *services.SlotsDoBetFreeReq) (resp *services.SlotsDoBetResp, err error) {
	resp = &services.SlotsDoBetResp{Code: services.ErrorCode_OK, Ret: false}
	if req == nil {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "")
	}
	if code, msg := validateSlotsFacadeBase(req.UserId, req.AgentId, req.GameId, req.RoundId, req.CurrencyType, req.FlowControl, req.GameStorage); code != services.ErrorCode_OK {
		return failSlotsDoBet(resp, code, msg)
	}

	win, winOK := parseAmountStrict(req.WinAmount)
	if !winOK {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "invalid winAmount")
	}
	if win.IsNegative() {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "winAmount can not be negative")
	}
	totalBet, betOK := parseAmountStrict(req.TotalBetAmount)
	if !betOK {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "invalid totalBetAmount")
	}
	if totalBet.IsNegative() {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "totalBetAmount can not be negative")
	}

	hasRecord := strings.TrimSpace(req.RecordJson) != ""
	if !hasRecord && win.GreaterThan(decimal.Zero) {
		return failSlotsDoBet(resp, services.ErrorCode_PARAMS_INVALID, "recordless winAmount must be zero")
	}

	// 免费中间步可能同 roundId 多次调用，仅 Complete=true 启用幂等。
	idemKey := ""
	idemSig := ""
	if req.Complete {
		idemKey = buildIdempotencyKey("slotsDoBetFree", u32Str(req.UserId), u32Str(req.GameId), req.CurrencyType, req.RoundId)
		idemSig = idempotencySignature(win.String(), totalBet.String(), req.RecordJson, req.Account, "complete")
		if hit, payload, code := d.beginIdempotency(idemKey, idemSig); code != services.ErrorCode_OK {
			msg := ""
			if code == services.ErrorCode_PARAMS_INVALID {
				msg = "idempotencyConflict"
			}
			return failSlotsDoBet(resp, code, msg)
		} else if hit {
			if cached, rCode := restoreSlotsDoBetResp(payload); rCode != services.ErrorCode_OK {
				return failSlotsDoBet(resp, rCode, "")
			} else {
				return cached, nil
			}
		}
	}

	currency := ""
	// Complete 只由 req.Complete 控制，不能用 hasRecord/win 代替。
	if req.Complete {
		slotsResp, callErr := d.SlotsLottery(ctx, &slotsLotteryReq{
			PlayerId:     req.UserId,
			CurrencyType: req.CurrencyType,
			AgentId:      int64(req.AgentId),
			GameId:       req.GameId,
			ProfitLoss:   win.String(),
			Bet:          "0",
			State:        req.RecordJson,
			RoundID:      req.RoundId,
			Complete:     true,
			Account:      req.Account,
		})
		if callErr != nil {
			d.abortIdempotency(idemKey)
			zap.L().Error("SlotsDoBetFree call SlotsLottery failed",
				zap.Uint32("userId", req.UserId),
				zap.String("roundId", req.RoundId),
				zap.Error(callErr))
			return failSlotsDoBet(resp, services.ErrorCode_SYSTEM_ERROR, "")
		}
		if slotsResp == nil {
			d.abortIdempotency(idemKey)
			return failSlotsDoBet(resp, services.ErrorCode_SYSTEM_ERROR, "")
		}
		if !slotsResp.Result || slotsResp.Code != services.ErrorCode_OK {
			d.abortIdempotency(idemKey)
			resp.Code = slotsResp.Code
			resp.Ret = false
			resp.Error = slotErrorName(slotsResp.Code)
			if slotsResp.NewCurrency != "" {
				resp.Currency, resp.CurrencyCent = fillBalanceResp(slotsResp.NewCurrency)
			}
			return resp, nil
		}
		currency = slotsResp.NewCurrency
	} else if win.GreaterThan(decimal.Zero) || hasRecord {
		nc, code := d.slotsFreeAwardWithoutComplete(req, win, totalBet)
		if code != services.ErrorCode_OK {
			return failSlotsDoBet(resp, code, "")
		}
		currency = nc.Truncate(2).StringFixed(2)
	} else {
		cent, code := d.getPlayerCurrency(req.UserId)
		if code != services.ErrorCode_OK {
			return failSlotsDoBet(resp, code, "")
		}
		currency = currencyFromCent(cent).Truncate(2).StringFixed(2)
	}

	var deleteKeys []string
	if req.FlowControl != nil {
		deleteKeys = req.FlowControl.DeleteGameStorageKeys
		if code := d.refundFacadePool(req.AgentId, req.UserId, req.GameId, req.CurrencyType, req.FlowControl.RefundPoolAmount); code != services.ErrorCode_OK {
			if idemKey != "" {
				d.abortIdempotency(idemKey)
			}
			return failSlotsDoBet(resp, code, "")
		}
	}
	if code := d.applyGameStorageWrites(req.UserId, req.GameId, req.CurrencyType, req.GameStorage, deleteKeys); code != services.ErrorCode_OK {
		if idemKey != "" {
			d.abortIdempotency(idemKey)
		}
		return failSlotsDoBet(resp, code, "")
	}

	resp.Ret = true
	resp.Currency, resp.CurrencyCent = fillBalanceResp(currency)
	if idemKey != "" {
		if raw, mErr := jsoniter.MarshalToString(resp); mErr == nil {
			d.commitIdempotency(idemKey, idemSig, raw)
		} else {
			d.abortIdempotency(idemKey)
		}
	}
	zap.L().Debug("SlotsDoBetFree success",
		zap.Uint32("userId", req.UserId),
		zap.Uint32("agentId", req.AgentId),
		zap.Uint32("gameId", req.GameId),
		zap.String("roundId", req.RoundId),
		zap.Bool("complete", req.Complete),
		zap.String("currency", resp.Currency))
	return resp, nil
}

// slotsFreeAwardWithoutComplete 免费局中间步骤：只返奖/写 ES 注单，不调用 CacheIns().Complete。
func (d *LotteryService) slotsFreeAwardWithoutComplete(req *services.SlotsDoBetFreeReq, win, totalBet decimal.Decimal) (decimal.Decimal, services.ErrorCode) {
	eAgent := dao.AgentManagerIns().Get(int64(req.AgentId))
	eGame := dao.GamesManagerIns().GetById(int64(req.GameId))
	if eAgent == nil {
		return decimal.Zero, services.ErrorCode_AGENT_FROZEN
	}
	if eGame == nil {
		return decimal.Zero, services.ErrorCode_PARAMS_INVALID
	}

	var newCurrency int64
	var code services.ErrorCode
	if win.GreaterThan(decimal.Zero) {
		newCurrency, code = d.updatePlayerCurrency(req.UserId, win.Mul(decimal.NewFromInt(100)).IntPart())
		if code != services.ErrorCode_OK {
			return decimal.Zero, code
		}
	} else {
		newCurrency, code = d.getPlayerCurrency(req.UserId)
		if code != services.ErrorCode_OK {
			return decimal.Zero, code
		}
	}

	nc := currencyFromCent(newCurrency)
	user := dao.CacheIns().GetUser(int64(req.AgentId), int64(req.UserId))
	if user != nil && user.IsTourist == 0 {
		account := req.Account
		if account == "" {
			account = user.Account
		}
		record := ConvertRecord(
			req.AgentId,
			req.UserId,
			req.RoundId,
			req.CurrencyType,
			eGame.ConfName,
			account,
			req.RecordJson,
			nc,
			uint32(eAgent.WebId),
			false,
			totalBet.InexactFloat64(),
			win.InexactFloat64(),
		)
		d.SaveRecord(record)
		if win.GreaterThan(decimal.Zero) {
			d.SaveBill(req.AgentId, req.UserId, win, nc.Truncate(2).InexactFloat64(), eGame.ConfName, "返奖", req.CurrencyType, req.RoundId)
		}
	}
	return nc, services.ErrorCode_OK
}

// FruitDoBet 单人 Fruit 门面：复用 doSingleBet（规则 A：水池/注单/Complete 均走原逻辑）。
// Complete 仅透传 req.Complete，由调用方显式标注。
func (d *LotteryService) FruitDoBet(ctx context.Context, req *services.FruitDoBetReq) (resp *services.FruitDoBetResp, err error) {
	resp = &services.FruitDoBetResp{Code: services.ErrorCode_OK, Ret: false}
	if req == nil {
		return failFruitDoBet(resp, services.ErrorCode_PARAMS_INVALID)
	}
	if code, _ := validateFacadeIdentity(req.UserId, req.AgentId, req.GameId, req.RoundId, req.CurrencyType); code != services.ErrorCode_OK {
		return failFruitDoBet(resp, code)
	}

	bet, betOK := parseAmountStrict(req.BetAmount)
	if !betOK {
		return failFruitDoBet(resp, services.ErrorCode_PARAMS_INVALID)
	}
	if bet.IsNegative() {
		return failFruitDoBet(resp, services.ErrorCode_PARAMS_INVALID)
	}
	win, winOK := parseAmountStrict(req.WinAmount)
	if !winOK {
		return failFruitDoBet(resp, services.ErrorCode_PARAMS_INVALID)
	}
	if win.IsNegative() {
		return failFruitDoBet(resp, services.ErrorCode_PARAMS_INVALID)
	}

	completeFlag := "0"
	if req.Complete {
		completeFlag = "1"
	}
	idemKey := buildIdempotencyKey("fruitDoBet", u32Str(req.UserId), u32Str(req.GameId), req.CurrencyType, req.RoundId)
	idemSig := idempotencySignature(bet.String(), win.String(), req.RecordJson, completeFlag)
	if hit, payload, code := d.beginIdempotency(idemKey, idemSig); code != services.ErrorCode_OK {
		return failFruitDoBet(resp, code)
	} else if hit {
		if cached, rCode := restoreFruitDoBetResp(payload); rCode != services.ErrorCode_OK {
			return failFruitDoBet(resp, rCode)
		} else {
			return cached, nil
		}
	}

	betResp, callErr := d.doSingleBet(&singleBetReq{
		UserId:       req.UserId,
		GameId:       req.GameId,
		Win:          win.String(),
		RoundID:      req.RoundId,
		Result:       req.RecordJson,
		Complete:     req.Complete,
		Bet:          bet.String(),
		AgentId:      req.AgentId,
		CurrencyType: req.CurrencyType,
	})
	if callErr != nil {
		d.abortIdempotency(idemKey)
		zap.L().Error("FruitDoBet call doSingleBet failed",
			zap.Uint32("userId", req.UserId),
			zap.String("roundId", req.RoundId),
			zap.Error(callErr))
		return failFruitDoBet(resp, services.ErrorCode_SYSTEM_ERROR)
	}
	if betResp == nil {
		d.abortIdempotency(idemKey)
		return failFruitDoBet(resp, services.ErrorCode_SYSTEM_ERROR)
	}
	resp.Code = betResp.Code
	if betResp.Currency != "" {
		resp.Currency, resp.CurrencyCent = fillBalanceResp(betResp.Currency)
	}
	if betResp.Code == services.ErrorCode_OK {
		resp.Ret = true
		if raw, mErr := jsoniter.MarshalToString(resp); mErr == nil {
			d.commitIdempotency(idemKey, idemSig, raw)
		} else {
			d.abortIdempotency(idemKey)
		}
		zap.L().Debug("FruitDoBet success",
			zap.Uint32("userId", req.UserId),
			zap.Uint32("agentId", req.AgentId),
			zap.Uint32("gameId", req.GameId),
			zap.String("roundId", req.RoundId),
			zap.Bool("complete", req.Complete),
			zap.String("currency", resp.Currency))
	} else {
		d.abortIdempotency(idemKey)
	}
	return resp, nil
}

// FruitDoBetMulti 百人扣款门面。
// 规则 A：复用 doMultiBet → deductBet（扣余额 + 改水池 + 流水）。
func (d *LotteryService) FruitDoBetMulti(ctx context.Context, req *services.FruitDoBetMultiReq) (resp *services.FruitDoBetMultiResp, err error) {
	resp = &services.FruitDoBetMultiResp{Code: services.ErrorCode_OK, Ret: false}
	if req == nil {
		return failFruitDoBetMulti(resp, services.ErrorCode_PARAMS_INVALID)
	}
	if code, _ := validateFacadeIdentity(req.UserId, req.AgentId, req.GameId, req.RoundId, req.CurrencyType); code != services.ErrorCode_OK {
		return failFruitDoBetMulti(resp, code)
	}

	bet, betOK := parseAmountStrict(req.BetAmount)
	if !betOK {
		return failFruitDoBetMulti(resp, services.ErrorCode_PARAMS_INVALID)
	}
	if !bet.GreaterThan(decimal.Zero) {
		return failFruitDoBetMulti(resp, services.ErrorCode_PARAMS_INVALID)
	}

	idemKey := buildIdempotencyKey("fruitDoBetMulti", u32Str(req.UserId), u32Str(req.GameId), req.CurrencyType, req.RoundId)
	idemSig := idempotencySignature(bet.String(), u32Str(req.AreaId))
	if hit, payload, code := d.beginIdempotency(idemKey, idemSig); code != services.ErrorCode_OK {
		return failFruitDoBetMulti(resp, code)
	} else if hit {
		if cached, rCode := restoreFruitDoBetMultiResp(payload); rCode != services.ErrorCode_OK {
			return failFruitDoBetMulti(resp, rCode)
		} else {
			return cached, nil
		}
	}

	betResp, callErr := d.doMultiBet(&multiBetReq{
		UserId:       req.UserId,
		GameId:       req.GameId,
		RoundID:      req.RoundId,
		AreaId:       req.AreaId,
		InitBet:      bet.String(),
		AgentId:      req.AgentId,
		CurrencyType: req.CurrencyType,
	})
	if callErr != nil {
		d.abortIdempotency(idemKey)
		zap.L().Error("FruitDoBetMulti call doMultiBet failed",
			zap.Uint32("userId", req.UserId),
			zap.String("roundId", req.RoundId),
			zap.Error(callErr))
		return failFruitDoBetMulti(resp, services.ErrorCode_SYSTEM_ERROR)
	}
	if betResp == nil {
		d.abortIdempotency(idemKey)
		return failFruitDoBetMulti(resp, services.ErrorCode_SYSTEM_ERROR)
	}
	// 在 deductBet 失败时可能仍返回 OK 且 currency 为空，门面按失败处理。
	if betResp.Code != services.ErrorCode_OK {
		d.abortIdempotency(idemKey)
		resp.Code = betResp.Code
		return resp, nil
	}
	if strings.TrimSpace(betResp.Currency) == "" {
		d.abortIdempotency(idemKey)
		zap.L().Error("FruitDoBetMulti empty currency after bet",
			zap.Uint32("userId", req.UserId),
			zap.String("roundId", req.RoundId),
			zap.String("bet", bet.String()))
		return failFruitDoBetMulti(resp, services.ErrorCode_SYSTEM_ERROR)
	}

	resp.Ret = true
	resp.Currency, resp.CurrencyCent = fillBalanceResp(betResp.Currency)
	if raw, mErr := jsoniter.MarshalToString(resp); mErr == nil {
		d.commitIdempotency(idemKey, idemSig, raw)
	} else {
		d.abortIdempotency(idemKey)
	}
	zap.L().Debug("FruitDoBetMulti success",
		zap.Uint32("userId", req.UserId),
		zap.Uint32("agentId", req.AgentId),
		zap.Uint32("gameId", req.GameId),
		zap.String("roundId", req.RoundId),
		zap.String("bet", bet.String()),
		zap.String("currency", resp.Currency))
	return resp, nil
}

// FruitRefundMulti 百人退款门面。
// 规则 A：复用 doMultiRefund → refundBet（退余额 + 回滚水池 + 流水）。
func (d *LotteryService) FruitRefundMulti(ctx context.Context, req *services.FruitRefundMultiReq) (resp *services.FruitRefundMultiResp, err error) {
	resp = &services.FruitRefundMultiResp{Code: services.ErrorCode_OK, Ret: false}
	if req == nil {
		return failFruitRefundMulti(resp, services.ErrorCode_PARAMS_INVALID)
	}
	if code, _ := validateFacadeIdentity(req.UserId, req.AgentId, req.GameId, req.RoundId, req.CurrencyType); code != services.ErrorCode_OK {
		return failFruitRefundMulti(resp, code)
	}

	bet, betOK := parseAmountStrict(req.BetAmount)
	if !betOK {
		return failFruitRefundMulti(resp, services.ErrorCode_PARAMS_INVALID)
	}
	if !bet.GreaterThan(decimal.Zero) {
		return failFruitRefundMulti(resp, services.ErrorCode_PARAMS_INVALID)
	}

	idemKey := buildIdempotencyKey("fruitRefundMulti", u32Str(req.UserId), u32Str(req.GameId), req.CurrencyType, req.RoundId)
	idemSig := idempotencySignature(bet.String())
	if hit, payload, code := d.beginIdempotency(idemKey, idemSig); code != services.ErrorCode_OK {
		return failFruitRefundMulti(resp, code)
	} else if hit {
		if cached, rCode := restoreFruitRefundMultiResp(payload); rCode != services.ErrorCode_OK {
			return failFruitRefundMulti(resp, rCode)
		} else {
			return cached, nil
		}
	}

	betResp, callErr := d.doMultiRefund(&multiRefundReq{
		UserId:       req.UserId,
		GameId:       req.GameId,
		Bet:          bet.String(),
		RoundID:      req.RoundId,
		AgentId:      req.AgentId,
		CurrencyType: req.CurrencyType,
	})
	if callErr != nil {
		d.abortIdempotency(idemKey)
		zap.L().Error("FruitRefundMulti call doMultiRefund failed",
			zap.Uint32("userId", req.UserId),
			zap.String("roundId", req.RoundId),
			zap.Error(callErr))
		return failFruitRefundMulti(resp, services.ErrorCode_SYSTEM_ERROR)
	}
	if betResp == nil {
		d.abortIdempotency(idemKey)
		return failFruitRefundMulti(resp, services.ErrorCode_SYSTEM_ERROR)
	}
	if betResp.Code != services.ErrorCode_OK {
		d.abortIdempotency(idemKey)
		resp.Code = betResp.Code
		return resp, nil
	}
	if strings.TrimSpace(betResp.Currency) == "" {
		d.abortIdempotency(idemKey)
		zap.L().Error("FruitRefundMulti empty currency after refund",
			zap.Uint32("userId", req.UserId),
			zap.String("roundId", req.RoundId),
			zap.String("bet", bet.String()))
		return failFruitRefundMulti(resp, services.ErrorCode_SYSTEM_ERROR)
	}

	resp.Ret = true
	resp.Currency, resp.CurrencyCent = fillBalanceResp(betResp.Currency)
	if raw, mErr := jsoniter.MarshalToString(resp); mErr == nil {
		d.commitIdempotency(idemKey, idemSig, raw)
	} else {
		d.abortIdempotency(idemKey)
	}
	zap.L().Debug("FruitRefundMulti success",
		zap.Uint32("userId", req.UserId),
		zap.Uint32("agentId", req.AgentId),
		zap.Uint32("gameId", req.GameId),
		zap.String("roundId", req.RoundId),
		zap.String("bet", bet.String()),
		zap.String("currency", resp.Currency))
	return resp, nil
}

// FruitSettleRound 百人整局结算门面。
// 规则 A：复用 doMultiSettle（入账模型为 win+bet；水池/注单/Complete 走原逻辑）。
func (d *LotteryService) FruitSettleRound(ctx context.Context, req *services.FruitSettleRoundReq) (resp *services.FruitSettleRoundResp, err error) {
	resp = &services.FruitSettleRoundResp{
		Code:    services.ErrorCode_OK,
		Ret:     false,
		Players: make([]*services.FruitSettlePlayerResult, 0),
	}
	if req == nil || len(req.Players) == 0 {
		return failFruitSettleRound(resp, services.ErrorCode_PARAMS_INVALID)
	}
	if code, _ := validateFacadeRoom(req.AgentId, req.GameId, req.RoundId, req.CurrencyType); code != services.ErrorCode_OK {
		return failFruitSettleRound(resp, code)
	}
	if req.Period < 0 {
		return failFruitSettleRound(resp, services.ErrorCode_PARAMS_INVALID)
	}

	records := make([]*settleRecord, 0, len(req.Players))
	totalWin := decimal.Zero
	seen := make(map[uint32]struct{}, len(req.Players))
	for _, p := range req.Players {
		if p == nil || p.UserId == 0 {
			return failFruitSettleRound(resp, services.ErrorCode_PARAMS_INVALID)
		}
		if _, dup := seen[p.UserId]; dup {
			zap.L().Error("FruitSettleRound duplicate user",
				zap.Uint32("userId", p.UserId),
				zap.String("roundId", req.RoundId))
			return failFruitSettleRound(resp, services.ErrorCode_PARAMS_INVALID)
		}
		seen[p.UserId] = struct{}{}

		win, winOK := parseAmountStrict(p.WinAmount)
		if !winOK || win.IsNegative() {
			return failFruitSettleRound(resp, services.ErrorCode_PARAMS_INVALID)
		}
		bet, betOK := parseAmountStrict(p.TotalBet)
		if !betOK || bet.IsNegative() {
			return failFruitSettleRound(resp, services.ErrorCode_PARAMS_INVALID)
		}
		totalWin = totalWin.Add(win)

		records = append(records, &settleRecord{
			UserId:       p.UserId,
			GameId:       req.GameId,
			Win:          win.String(),
			RoundID:      req.RoundId,
			Log:          p.RecordJson,
			Bet:          bet.String(),
			CurrencyType: req.CurrencyType,
			AgentId:      req.AgentId,
			Account:      p.Account,
		})
	}

	idemKey := buildIdempotencyKey("fruitSettleRound", u32Str(req.AgentId), u32Str(req.GameId), req.CurrencyType, i64Str(req.Period), req.RoundId)
	sigParts := make([]string, 0, len(records)*3+1)
	sigParts = append(sigParts, totalWin.String())
	for _, r := range records {
		sigParts = append(sigParts, u32Str(r.UserId), r.Bet, r.Win)
	}
	idemSig := idempotencySignature(sigParts...)
	if hit, payload, code := d.beginIdempotency(idemKey, idemSig); code != services.ErrorCode_OK {
		return failFruitSettleRound(resp, code)
	} else if hit {
		if cached, rCode := restoreFruitSettleRoundResp(payload); rCode != services.ErrorCode_OK {
			return failFruitSettleRound(resp, rCode)
		} else {
			return cached, nil
		}
	}

	betResp, callErr := d.doMultiSettle(&multiSettleReq{
		Records:  records,
		TotalWin: totalWin.String(),
	})
	if callErr != nil {
		d.abortIdempotency(idemKey)
		zap.L().Error("FruitSettleRound call doMultiSettle failed",
			zap.Uint32("agentId", req.AgentId),
			zap.Uint32("gameId", req.GameId),
			zap.String("roundId", req.RoundId),
			zap.Error(callErr))
		return failFruitSettleRound(resp, services.ErrorCode_SYSTEM_ERROR)
	}
	if betResp == nil {
		d.abortIdempotency(idemKey)
		return failFruitSettleRound(resp, services.ErrorCode_SYSTEM_ERROR)
	}
	resp.Code = betResp.Code
	if betResp.Code != services.ErrorCode_OK {
		d.abortIdempotency(idemKey)
		return resp, nil
	}

	for _, item := range betResp.Currencys {
		if item == nil {
			continue
		}
		currency, cent := fillBalanceResp(item.Currency)
		resp.Players = append(resp.Players, &services.FruitSettlePlayerResult{
			UserId:       item.UserId,
			Currency:     currency,
			CurrencyCent: cent,
		})
	}
	resp.Ret = true
	if raw, mErr := jsoniter.MarshalToString(resp); mErr == nil {
		d.commitIdempotency(idemKey, idemSig, raw)
	} else {
		d.abortIdempotency(idemKey)
	}
	zap.L().Debug("FruitSettleRound success",
		zap.Uint32("agentId", req.AgentId),
		zap.Uint32("gameId", req.GameId),
		zap.String("roundId", req.RoundId),
		zap.Int64("period", req.Period),
		zap.Int("playerCount", len(resp.Players)),
		zap.String("totalWin", totalWin.String()))
	return resp, nil
}
