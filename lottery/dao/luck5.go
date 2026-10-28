package dao

import (
	"app/config"
	"app/esindex"
	"app/poollog"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"math"
	"micro_service/services"
	"strconv"
	"time"
	// 发牌时固定资金身份、原池与汇率，结算不再按当前配置选择水池。
	"go.uber.org/zap"
)

type Luck5Round struct {
	PoolSymbol   string                   `json:"poolSymbol,omitempty"`
	PoolVersion  int64                    `json:"poolVersion,omitempty"`
	PoolBase     decimal.Decimal          `json:"poolBase"`
	UserId       uint32                   `json:"userId"`
	AgentId      uint32                   `json:"agentId"`
	GameId       uint32                   `json:"gameId"`
	CurrencyType string                   `json:"currencyType"`
	RoundId      string                   `json:"roundId"`
	BetCent      int64                    `json:"betCent"`
	ReserveCent  int64                    `json:"reserveCent"`
	Revenue      decimal.Decimal          `json:"revenue"`
	RevenueRate  decimal.Decimal          `json:"revenueRate"`
	Symbol       string                   `json:"symbol"`
	Exchange     decimal.Decimal          `json:"exchange"`
	Account      string                   `json:"account"`
	WebId        uint32                   `json:"webId"`
	GameName     string                   `json:"gameName"`
	IsTourist    int32                    `json:"isTourist"`
	DealBalance  int64                    `json:"dealBalance"`
	DealCode     services.ErrorCode       `json:"dealCode"`
	DealError    string                   `json:"dealError"`
	Signature    string                   `json:"signature"`
	AttemptId    string                   `json:"attemptId"`
	Attempts     map[string]*Luck5Attempt `json:"attempts,omitempty"`
	WinCent      int64                    `json:"winCent"`
	RecordJson   string                   `json:"recordJson"`
	Balance      int64                    `json:"balance"`
	FundsDone    bool                     `json:"fundsDone"`
	RecordDone   bool                     `json:"recordDone"`
}
type Luck5Attempt struct {
	Signature string             `json:"signature"`
	Code      services.ErrorCode `json:"code"`
	Error     string             `json:"error"`
	Balance   int64              `json:"balance"`
	Accepted  bool               `json:"accepted"`
}

// 局号全局唯一，不能让相同局号在另一个用户、代理或币种下再次扣注。
func luck5Key(roundId string) (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("roundId", roundId))
	poolLogTrace := poollog.BeginDebug("dao.luck5Key", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	digest := sha256.Sum256([]byte(roundId))
	return "luck5:round:" + hex.EncodeToString(digest[:])
}

// 钱包使用整数分；既有水池和统计按配置汇率换算到池单位。
func luck5Amount(cent int64) (poolLogResult0 decimal.Decimal) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("dao.luck5Amount", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return decimal.NewFromInt(cent).Shift(-2)
}
func luck5DealPoolAfter(pool, bet, revenue decimal.Decimal) (poolLogResult0 decimal.Decimal) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("dao.luck5DealPoolAfter", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return pool.Add(bet).Sub(revenue)
}
func luck5CanSettle(pool, reserve, win decimal.Decimal) (poolLogResult0 bool) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("dao.luck5CanSettle", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return pool.Add(reserve).GreaterThanOrEqual(win)
}

// 在代理锁内读取持久化阶段，网络结果未知时也能按原局号确认。
func (rd *RedisDao) loadLuck5Round(ctx context.Context, roundId string) (poolLogResult0 *Luck5Round, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(rd)...)
	poolLogFields = append(poolLogFields, zap.String("roundId", roundId))
	poolLogTrace := poollog.Begin("dao.RedisDao.loadLuck5Round", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	raw, err := rd.cli.Get(ctx, luck5Key(roundId)).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.loadLuck5Round", "rd.cli.Get.Result", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	round := &Luck5Round{}
	if err := jsoniter.UnmarshalFromString(raw, round); err != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.loadLuck5Round", "jsoniter.UnmarshalFromString", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	return round, nil
}

// 身份核对不能只看用户与局号，否则原池预扣可能被另一代理或币种使用。
func sameLuck5Identity(round, request *Luck5Round) (poolLogResult0 bool) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(round)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(request)...)
	poolLogTrace := poollog.BeginDebug("dao.sameLuck5Identity", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return round.UserId == request.UserId && round.AgentId == request.AgentId && round.GameId == request.GameId && round.CurrencyType == request.CurrencyType && round.RoundId == request.RoundId
}

// 重复发牌先返回原扣注结果，配置或玩家缓存变化不影响已成功的局。
func (rd *RedisDao) Luck5DealResult(ctx context.Context, request *Luck5Round) (poolLogResult0 *services.Luck5Resp, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(rd)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(request)...)
	poolLogTrace := poollog.Begin("dao.RedisDao.Luck5DealResult", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	round, err := rd.loadLuck5Round(ctx, request.RoundId)
	if err != nil || round == nil {
		if err != nil {
			poollog.Event(zap.ErrorLevel, "dao.RedisDao.Luck5DealResult", "rd.loadLuck5Round", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		}
		return nil, err
	}
	if !sameLuck5Identity(round, request) || round.BetCent != request.BetCent || round.ReserveCent != request.ReserveCent {
		return &services.Luck5Resp{Code: services.ErrorCode_PARAMS_INVALID, Error: "dealConflict"}, nil
	}
	return &services.Luck5Resp{Code: round.DealCode, Ret: round.DealCode == services.ErrorCode_OK, CurrencyCent: round.DealBalance, Error: round.DealError}, nil
}

// 资金事务同时提交钱包、原池统计和阶段结果；只在成功确认后推进内存缓存。
func (rd *RedisDao) commitLuck5Funds(ctx context.Context, agent *AgentData, round *Luck5Round, settle bool) (poolLogResult0 services.ErrorCode, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(rd)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(round)...)
	poolLogTrace := poollog.Begin("dao.RedisDao.commitLuck5Funds", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	playerKey := "player_" + strconv.FormatUint(uint64(round.UserId), 10)
	roundKey := luck5Key(round.RoundId)
	if round.PoolSymbol == "" {
		round.PoolSymbol = round.Symbol
	}
	game := agent.getPoolGame(round.PoolSymbol)
	user := agent.GetUser(round.UserId)
	nextGame := *game
	nextUser := *user
	bet := luck5Amount(round.BetCent).Mul(round.Exchange)
	win := luck5Amount(round.WinCent).Mul(round.Exchange)
	reserve := luck5Amount(round.ReserveCent).Mul(round.Exchange)
	revenue := round.Revenue
	if round.IsTourist == 0 {
		if settle {
			nextGame.TotalProfLoss = nextGame.TotalProfLoss.Sub(reserve.Sub(win))
			chips := decimal.Max(bet, win)
			nextGame.TotalChips = nextGame.TotalChips.Add(chips)
			nextUser.TotalProfLoss = nextUser.TotalProfLoss.Add(win)
			nextUser.Count = nextUser.Count.Add(decimal.NewFromInt(1))
		} else {
			nextGame.TotalEffectBet = nextGame.TotalEffectBet.Add(bet)
			nextGame.TotalRevenue = nextGame.TotalRevenue.Add(revenue)
			nextGame.TotalProfLoss = nextGame.TotalProfLoss.Add(reserve)
			nextUser.TotalEffectBet = nextUser.TotalEffectBet.Add(bet)
		}
	}
	code := services.ErrorCode_SYSTEM_ERROR
	storedRaw, readErr := rd.cli.Get(ctx, roundKey).Result()
	if readErr != nil && !errors.Is(readErr, redis.Nil) {
		if readErr != nil {
			poollog.Event(zap.ErrorLevel, "dao.RedisDao.commitLuck5Funds", "rd.cli.Get.Result", "水池步骤失败", append(poolLogFields, zap.Error(readErr))...)
		}
		return code, readErr
	}
	if settle {
		stored := &Luck5Round{}
		if readErr != nil || jsoniter.UnmarshalFromString(storedRaw, stored) != nil || stored.FundsDone || !sameLuck5Identity(stored, round) {
			if readErr != nil {
				poollog.Event(zap.ErrorLevel, "dao.RedisDao.commitLuck5Funds", "check", "水池步骤失败", append(poolLogFields, zap.Error(readErr))...)
			}
			return code, errors.New("round stage changed")
		}
	} else if readErr == nil {
		return code, errors.New("round already exists")
	}
	current, err := rd.cli.HGet(ctx, playerKey, "currency").Int64()
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.commitLuck5Funds", "rd.cli.HGet.Int64", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return code, err
	}
	if current < 0 {
		return code, errors.New("invalid player balance")
	}
	if settle {
		if round.WinCent > math.MaxInt64-current {
			code = services.ErrorCode_PARAMS_INVALID
			return code, errors.New("balance overflow")
		}
		round.Balance = current + round.WinCent
		round.FundsDone = true
	} else {
		if current < round.BetCent {
			code = services.ErrorCode_NO_ENOUGH_MONEY
			return code, errors.New("balanceInsufficient")
		}
		round.DealBalance = current - round.BetCent
		round.Balance = round.DealBalance
	}
	raw, err := jsoniter.MarshalToString(round)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.commitLuck5Funds", "jsoniter.MarshalToString", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return code, err
	}
	_, err = rd.cli.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.HSet(ctx, playerKey, "currency", round.Balance)
		pipe.Expire(ctx, playerKey, 20*time.Minute)
		pipe.SAdd(ctx, "dirty_list_imp", round.UserId)
		pipe.Set(ctx, roundKey, raw, 0)
		if round.IsTourist == 0 {
			member := fmt.Sprintf("%d_%s", agent.Id, round.PoolSymbol)
			userMember := strconv.FormatUint(uint64(round.UserId), 10)
			if settle {
				pipe.ZIncrBy(ctx, esindex.AgentProfitLossData(), win.Sub(reserve).InexactFloat64(), member)
				pipe.ZIncrBy(ctx, esindex.AgentChipsData(), decimal.Max(bet, win).InexactFloat64(), member)
				pipe.ZIncrBy(ctx, "userTotalProfLoss", win.InexactFloat64(), userMember)
				pipe.ZIncrBy(ctx, "userBetCount", 1, userMember)
			} else {
				pipe.ZIncrBy(ctx, esindex.AgentEffectData(), bet.InexactFloat64(), member)
				pipe.ZIncrBy(ctx, esindex.AgentRevenueData(), revenue.InexactFloat64(), member)
				pipe.ZIncrBy(ctx, esindex.AgentProfitLossData(), reserve.InexactFloat64(), member)
				pipe.ZIncrBy(ctx, "userTotalEffBet", bet.InexactFloat64(), userMember)
			}
		}
		return nil
	})
	if err == nil {
		code = services.ErrorCode_OK
	}
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.commitLuck5Funds", "rd.cli.TxPipelined", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		// EXEC 回包丢失时确认持久化阶段；不盲目再扣钱包或恢复旧池快照。
		stored, readErr := rd.loadLuck5Round(context.Background(), round.RoundId)
		if readErr != nil || stored == nil || !sameLuck5Identity(stored, round) || stored.BetCent != round.BetCent || stored.ReserveCent != round.ReserveCent || (settle && (!stored.FundsDone || stored.Signature != round.Signature)) || (!settle && stored.FundsDone) {
			if readErr != nil {
				poollog.Event(zap.ErrorLevel, "dao.RedisDao.commitLuck5Funds", "rd.loadLuck5Round", "水池步骤失败", append(poolLogFields, zap.Error(readErr))...)
			}
			delete(agent.gameCache, round.PoolSymbol)
			delete(agent.userCache, round.UserId)
			return code, err
		}
		*round = *stored
	}
	nextGame.UpdateTime = time.Now().Unix()
	nextUser.UpdateTime = nextGame.UpdateTime
	*game = nextGame
	*user = nextUser
	return services.ErrorCode_OK, nil
}

// 发牌的余额检查、池检查和资金更新都在既有代理锁内完成。
func (gcm *GameCacheMgr) Luck5Deal(ctx context.Context, request *Luck5Round) (poolLogResult0 *services.Luck5Resp, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(gcm)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(request)...)
	poolLogTrace := poollog.Begin("dao.GameCacheMgr.Luck5Deal", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	agent := gcm.GetAgent(int64(request.AgentId))
	agent.lock.Lock()
	defer agent.lock.Unlock()
	if response, err := RedisIns().Luck5DealResult(ctx, request); err != nil || response != nil {
		if err != nil {
			poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Deal", "RedisIns.Luck5DealResult", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		}
		return response, err
	}
	balance, err := RedisIns().GetPlayerCurrency(request.UserId)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Deal", "RedisIns.GetPlayerCurrency", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	if balance < request.BetCent {
		return RedisIns().saveLuck5DealRejection(ctx, request, &services.Luck5Resp{Code: services.ErrorCode_NO_ENOUGH_MONEY, CurrencyCent: balance, Error: "balanceInsufficient"})
	}
	game := agent.GetGame(request.Symbol)
	request.PoolSymbol = game.Symbol
	request.PoolBase = poolCfgBase(agent.Id, request.Symbol)
	if group := config.CfgIns.GetPoolGroup(agent.Id, request.Symbol); group != nil {
		request.PoolVersion = group.Version
	}
	user := agent.GetUser(request.UserId)
	request.Account = user.Account
	request.IsTourist = user.IsTourist
	bet := luck5Amount(request.BetCent).Mul(request.Exchange)
	reserve := luck5Amount(request.ReserveCent).Mul(request.Exchange)
	request.Revenue = bet.Mul(request.RevenueRate).Truncate(4)
	currentPool := poolWithBase(game, agent.Id, request.Symbol)
	if user.IsTourist == 0 && luck5DealPoolAfter(currentPool, bet, request.Revenue).LessThan(reserve) {
		return RedisIns().saveLuck5DealRejection(ctx, request, &services.Luck5Resp{Code: services.ErrorCode_NO_ENOUGH_POOL_MONEY, CurrencyCent: balance, Error: "poolInsufficient"})
	}
	request.DealCode = services.ErrorCode_OK
	request.DealError = ""
	if code, err := RedisIns().commitLuck5Funds(ctx, agent, request, false); err != nil {
		poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Deal", "RedisIns.commitLuck5Funds", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return &services.Luck5Resp{Code: code, CurrencyCent: balance, Error: "dealFundsFailed"}, err
	}
	return &services.Luck5Resp{Code: services.ErrorCode_OK, Ret: true, CurrencyCent: request.DealBalance}, nil
}
func (rd *RedisDao) saveLuck5DealRejection(ctx context.Context, request *Luck5Round, response *services.Luck5Resp) (poolLogResult0 *services.Luck5Resp, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(rd)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(request)...)
	poolLogTrace := poollog.Begin("dao.RedisDao.saveLuck5DealRejection", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	request.DealCode = response.Code
	request.DealError = response.Error
	request.DealBalance = response.CurrencyCent
	raw, err := jsoniter.MarshalToString(request)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.saveLuck5DealRejection", "jsoniter.MarshalToString", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	if err := rd.cli.SetNX(ctx, luck5Key(request.RoundId), raw, 0).Err(); err != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.saveLuck5DealRejection", "rd.cli.SetNX.Err", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		if !errors.Is(err, redis.Nil) {
			return nil, err
		}
	}
	stored, loadErr := rd.loadLuck5Round(ctx, request.RoundId)
	if loadErr != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.saveLuck5DealRejection", "rd.loadLuck5Round", "水池步骤失败", append(poolLogFields, zap.Error(loadErr))...)
		return nil, loadErr
	}
	if stored == nil || !sameLuck5Identity(stored, request) || stored.BetCent != request.BetCent || stored.ReserveCent != request.ReserveCent {
		return &services.Luck5Resp{Code: services.ErrorCode_PARAMS_INVALID, Error: "dealConflict"}, nil
	}
	return &services.Luck5Resp{Code: stored.DealCode, Ret: stored.DealCode == services.ErrorCode_OK, CurrencyCent: stored.DealBalance, Error: stored.DealError}, nil
}

// 结算只动已预扣资金；资金阶段已提交时重试只保存原记录。
func (gcm *GameCacheMgr) Luck5Settle(ctx context.Context, request *Luck5Round, saveRecord func(*Luck5Round) error) (poolLogResult0 *services.Luck5Resp, poolLogResult1 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(gcm)...)
	poolLogFields = append(poolLogFields, poollog.MetadataFields(request)...)
	poolLogTrace := poollog.Begin("dao.GameCacheMgr.Luck5Settle", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult1)
	agent := gcm.GetAgent(int64(request.AgentId))
	agent.lock.Lock()
	defer agent.lock.Unlock()
	round, err := RedisIns().loadLuck5Round(ctx, request.RoundId)
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Settle", "RedisIns.loadLuck5Round", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return nil, err
	}
	if round == nil || !sameLuck5Identity(round, request) {
		return &services.Luck5Resp{Code: services.ErrorCode_PARAMS_INVALID, Error: "roundIdentityMismatch"}, nil
	}
	if request.WinCent < 0 || request.AttemptId == "" {
		return &services.Luck5Resp{Code: services.ErrorCode_PARAMS_INVALID, Error: "invalidAttempt"}, nil
	}
	if round.Attempts == nil {
		round.Attempts = make(map[string]*Luck5Attempt)
	}
	if attempt := round.Attempts[request.AttemptId]; attempt != nil {
		if attempt.Signature != request.Signature {
			return &services.Luck5Resp{Code: services.ErrorCode_PARAMS_INVALID, Error: "settleConflict"}, nil
		}
		if !attempt.Accepted {
			return &services.Luck5Resp{Code: attempt.Code, Ret: false, CurrencyCent: attempt.Balance, Error: attempt.Error}, nil
		}
	} else if round.FundsDone || round.AttemptId != "" {
		return &services.Luck5Resp{Code: services.ErrorCode_PARAMS_INVALID, Error: "settleAlreadyAccepted"}, nil
	}
	if round.RecordDone {
		return &services.Luck5Resp{Code: services.ErrorCode_OK, Ret: true, CurrencyCent: round.Balance}, nil
	}
	if !round.FundsDone {
		if round.PoolSymbol == "" {
			round.PoolSymbol = round.Symbol
			round.PoolBase = poolCfgBase(agent.Id, round.Symbol)
		}
		agentGame := agent.getPoolGame(round.PoolSymbol)
		pool := poolRaw(agentGame).Add(round.PoolBase)
		reserve := luck5Amount(round.ReserveCent).Mul(round.Exchange)
		win := luck5Amount(request.WinCent).Mul(round.Exchange)
		if round.IsTourist == 0 && !luck5CanSettle(pool, reserve, win) {
			balance, balanceErr := RedisIns().GetPlayerCurrency(round.UserId)
			if balanceErr != nil {
				poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Settle", "RedisIns.GetPlayerCurrency", "水池步骤失败", append(poolLogFields, zap.Error(balanceErr))...)
				return &services.Luck5Resp{Code: services.ErrorCode_SYSTEM_ERROR, Error: "balanceReadFailed"}, balanceErr
			}
			attempt := &Luck5Attempt{Signature: request.Signature, Code: services.ErrorCode_NO_ENOUGH_POOL_MONEY, Error: "poolInsufficient", Balance: balance}
			round.Attempts[request.AttemptId] = attempt
			raw, marshalErr := jsoniter.MarshalToString(round)
			if marshalErr != nil {
				poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Settle", "jsoniter.MarshalToString", "水池步骤失败", append(poolLogFields, zap.Error(marshalErr))...)
				return nil, marshalErr
			}
			if err := RedisIns().cli.Set(ctx, luck5Key(round.RoundId), raw, 0).Err(); err != nil {
				poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Settle", "RedisIns.cli.Set.Err", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
				return &services.Luck5Resp{Code: services.ErrorCode_SYSTEM_ERROR, Error: "attemptSaveFailed"}, err
			}
			return &services.Luck5Resp{Code: services.ErrorCode_NO_ENOUGH_POOL_MONEY, CurrencyCent: balance, Error: "poolInsufficient"}, nil
		}
		round.WinCent = request.WinCent
		round.RecordJson = request.RecordJson
		round.Signature = request.Signature
		round.AttemptId = request.AttemptId
		round.Attempts[request.AttemptId] = &Luck5Attempt{Signature: request.Signature, Code: services.ErrorCode_OK, Accepted: true}
		if code, err := RedisIns().commitLuck5Funds(ctx, agent, round, true); err != nil {
			poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Settle", "RedisIns.commitLuck5Funds", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return &services.Luck5Resp{Code: code, Error: "settleFundsFailed"}, err
		}
	}
	if err := saveRecord(round); err != nil {
		poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Settle", "saveRecord", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return &services.Luck5Resp{Code: services.ErrorCode_SYSTEM_ERROR, CurrencyCent: round.Balance, Error: "recordSaveFailed"}, err
	}
	round.RecordDone = true
	raw, err := jsoniter.MarshalToString(round)
	if err == nil {
		err = RedisIns().cli.Set(ctx, luck5Key(round.RoundId), raw, 0).Err()
	}
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.GameCacheMgr.Luck5Settle", "jsoniter.MarshalToString", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return &services.Luck5Resp{Code: services.ErrorCode_SYSTEM_ERROR, CurrencyCent: round.Balance, Error: "completionSaveFailed"}, err
	}
	return &services.Luck5Resp{Code: services.ErrorCode_OK, Ret: true, CurrencyCent: round.Balance}, nil
}
