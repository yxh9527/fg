package dao

import (
	"app/esindex"
	"app/poollog"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"micro_service/services"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Luck5Round是未结算局的原下注、预占和原池参数；完成后删除，不保存防重结果。
type Luck5Round struct {
	PoolSymbol   string
	PoolBase     decimal.Decimal
	UserId       uint32
	AgentId      uint32
	GameId       uint32
	CurrencyType string
	RoundId      string
	BetCent      int64
	ReserveCent  int64
	Revenue      decimal.Decimal
	RevenueRate  decimal.Decimal
	Symbol       string
	Exchange     decimal.Decimal
	Account      string
	WebId        uint32
	GameName     string
	IsTourist    int32
	DealBalance  int64
	WinCent      int64
	RecordJson   string
	Balance      int64
}

func luck5Amount(cent int64) (result decimal.Decimal) {
	poolLogTrace := poollog.BeginDebug("dao.luck5Amount")
	defer poolLogTrace.Finish(nil)
	return decimal.NewFromInt(cent).Shift(-2)
}
func luck5DealPoolAfter(pool, bet, revenue decimal.Decimal) (result decimal.Decimal) {
	poolLogTrace := poollog.BeginDebug("dao.luck5DealPoolAfter")
	defer poolLogTrace.Finish(nil)
	return pool.Add(bet).Sub(revenue)
}
func luck5CanSettle(pool, reserve, win decimal.Decimal) (result bool) {
	poolLogTrace := poollog.BeginDebug("dao.luck5CanSettle")
	defer poolLogTrace.Finish(nil)
	return pool.Add(reserve).GreaterThanOrEqual(win)
}

// 局号只用于关联Deal和Settle的业务参数，不查询或保存重复请求结果。
func luck5Key(roundID string) (result string) {
	poolLogTrace := poollog.BeginDebug("dao.luck5Key", zap.String("roundId", roundID))
	defer poolLogTrace.Finish(nil)
	sum := sha256.Sum256([]byte(roundID))
	return "luck5:round:" + hex.EncodeToString(sum[:])
}
func (rd *RedisDao) loadLuck5Round(ctx context.Context, roundID string) (round *Luck5Round, err error) {
	poolLogTrace := poollog.Begin("dao.RedisDao.loadLuck5Round", zap.DebugLevel, zap.String("roundId", roundID))
	defer poolLogTrace.Finish(&err)
	raw, err := rd.cli.Get(ctx, luck5Key(roundID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	round = &Luck5Round{}
	if err = json.Unmarshal(raw, round); err != nil {
		return nil, err
	}
	return round, nil
}

// 游戏服务器负责防重；lottery按本次请求更新钱包、水池及玩家累计。
func (rd *RedisDao) commitLuck5Funds(ctx context.Context, agent *AgentData, round *Luck5Round, settle bool) (code services.ErrorCode, err error) {
	poolLogTrace := poollog.Begin("dao.RedisDao.commitLuck5Funds", zap.DebugLevel, poollog.MetadataFields(round)...)
	defer poolLogTrace.Finish(&err)
	code = services.ErrorCode_SYSTEM_ERROR
	if round.BetCent < 0 || round.ReserveCent < 0 || round.WinCent < 0 {
		return services.ErrorCode_PARAMS_INVALID, errors.New("invalid amount")
	}
	playerKey := "player_" + strconv.FormatUint(uint64(round.UserId), 10)
	if round.PoolSymbol == "" {
		round.PoolSymbol = round.Symbol
	}
	game := agent.getPoolGame(round.PoolSymbol)
	user := agent.GetUser(round.UserId)
	nextGame, nextUser := *game, *user
	bet := luck5Amount(round.BetCent).Mul(round.Exchange)
	win := luck5Amount(round.WinCent).Mul(round.Exchange)
	reserve := luck5Amount(round.ReserveCent).Mul(round.Exchange)
	if round.IsTourist == 0 {
		if settle {
			nextGame.TotalProfLoss = nextGame.TotalProfLoss.Add(win.Sub(reserve))
			nextGame.TotalChips = nextGame.TotalChips.Add(decimal.Max(bet, win))
			nextUser.TotalProfLoss = nextUser.TotalProfLoss.Add(win)
			nextUser.Count = nextUser.Count.Add(decimal.NewFromInt(1))
		} else {
			nextGame.TotalEffectBet = nextGame.TotalEffectBet.Add(bet)
			nextGame.TotalRevenue = nextGame.TotalRevenue.Add(round.Revenue)
			nextGame.TotalProfLoss = nextGame.TotalProfLoss.Add(reserve)
			nextUser.TotalEffectBet = nextUser.TotalEffectBet.Add(bet)
		}
	}
	current, err := rd.cli.HGet(ctx, playerKey, "currency").Int64()
	if err != nil {
		return code, err
	}
	if current < 0 {
		return code, errors.New("invalid player balance")
	}
	if settle {
		if round.WinCent > math.MaxInt64-current {
			return services.ErrorCode_PARAMS_INVALID, errors.New("balance overflow")
		}
		round.Balance = current + round.WinCent
	} else {
		if current < round.BetCent {
			return services.ErrorCode_NO_ENOUGH_MONEY, errors.New("balanceInsufficient")
		}
		round.DealBalance = current - round.BetCent
		round.Balance = round.DealBalance
	}
	var contextJSON []byte
	if !settle {
		contextJSON, err = json.Marshal(round)
		if err != nil {
			return code, err
		}
	}
	_, err = rd.cli.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if !settle {
			pipe.Set(ctx, luck5Key(round.RoundId), contextJSON, time.Hour)
		}
		pipe.HSet(ctx, playerKey, "currency", round.Balance)
		pipe.Expire(ctx, playerKey, 20*time.Minute)
		pipe.SAdd(ctx, "dirty_list_imp", round.UserId)
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
				pipe.ZIncrBy(ctx, esindex.AgentRevenueData(), round.Revenue.InexactFloat64(), member)
				pipe.ZIncrBy(ctx, esindex.AgentProfitLossData(), reserve.InexactFloat64(), member)
				pipe.ZIncrBy(ctx, "userTotalEffBet", bet.InexactFloat64(), userMember)
			}
		}
		return nil
	})
	if err != nil {
		delete(agent.gameCache, round.PoolSymbol)
		delete(agent.userCache, round.UserId)
		return code, err
	}
	nextGame.UpdateTime = time.Now().Unix()
	nextUser.UpdateTime = nextGame.UpdateTime
	*game = nextGame
	*user = nextUser
	return services.ErrorCode_OK, nil
}

// 发牌只检查本次余额及可用水池，不查找历史局或保存拒绝记录。
func (gcm *GameCacheMgr) Luck5Deal(ctx context.Context, request *Luck5Round) (response *services.Luck5Resp, err error) {
	poolLogTrace := poollog.Begin("dao.GameCacheMgr.Luck5Deal", zap.DebugLevel, poollog.MetadataFields(request)...)
	defer poolLogTrace.Finish(&err)
	agent := gcm.GetAgent(int64(request.AgentId))
	agent.lock.Lock()
	defer agent.lock.Unlock()
	balance, err := RedisIns().GetPlayerCurrency(request.UserId)
	if err != nil {
		return nil, err
	}
	if balance < request.BetCent {
		return &services.Luck5Resp{Code: services.ErrorCode_NO_ENOUGH_MONEY, CurrencyCent: balance, Error: "balanceInsufficient"}, nil
	}
	game := agent.GetGame(request.Symbol)
	request.PoolSymbol = game.Symbol
	request.PoolBase = poolCfgBase(agent.Id, request.Symbol)
	user := agent.GetUser(request.UserId)
	request.Account = user.Account
	request.IsTourist = user.IsTourist
	bet := luck5Amount(request.BetCent).Mul(request.Exchange)
	reserve := luck5Amount(request.ReserveCent).Mul(request.Exchange)
	request.Revenue = bet.Mul(request.RevenueRate).Truncate(4)
	if user.IsTourist == 0 && luck5DealPoolAfter(poolWithBase(game, agent.Id, request.Symbol), bet, request.Revenue).LessThan(reserve) {
		return &services.Luck5Resp{Code: services.ErrorCode_NO_ENOUGH_POOL_MONEY, CurrencyCent: balance, Error: "poolInsufficient"}, nil
	}
	code, err := RedisIns().commitLuck5Funds(ctx, agent, request, false)
	if err != nil {
		return &services.Luck5Resp{Code: code, CurrencyCent: balance, Error: "dealFundsFailed"}, err
	}
	return &services.Luck5Resp{Code: services.ErrorCode_OK, Ret: true, CurrencyCent: request.DealBalance}, nil
}

// 本文件的方法只保存正常结算所需的Deal参数，不保存防重结果或候选状态。
func (gcm *GameCacheMgr) Luck5Settle(ctx context.Context, request *Luck5Round, saveRecord func(*Luck5Round) error) (response *services.Luck5Resp, err error) {
	poolLogTrace := poollog.Begin("dao.GameCacheMgr.Luck5Settle", zap.DebugLevel, poollog.MetadataFields(request)...)
	defer poolLogTrace.Finish(&err)
	agent := gcm.GetAgent(int64(request.AgentId))
	agent.lock.Lock()
	defer agent.lock.Unlock()
	round, err := RedisIns().loadLuck5Round(ctx, request.RoundId)
	if err != nil {
		return nil, err
	}
	if round == nil {
		return &services.Luck5Resp{Code: services.ErrorCode_PARAMS_INVALID, Error: "roundContextMissing"}, nil
	}
	if round.RoundId != request.RoundId || round.UserId != request.UserId || round.AgentId != request.AgentId || round.GameId != request.GameId || round.CurrencyType != request.CurrencyType {
		return &services.Luck5Resp{Code: services.ErrorCode_PARAMS_INVALID, Error: "roundIdentityMismatch"}, nil
	}
	if request.WinCent < 0 {
		return &services.Luck5Resp{Code: services.ErrorCode_PARAMS_INVALID, Error: "invalidAmount"}, nil
	}
	round.WinCent = request.WinCent
	round.RecordJson = request.RecordJson
	if round.PoolSymbol == "" {
		round.PoolSymbol = round.Symbol
	}
	game := agent.getPoolGame(round.PoolSymbol)
	reserve := luck5Amount(round.ReserveCent).Mul(round.Exchange)
	win := luck5Amount(round.WinCent).Mul(round.Exchange)
	if round.IsTourist == 0 && !luck5CanSettle(poolRaw(game).Add(round.PoolBase), reserve, win) {
		balance, readErr := RedisIns().GetPlayerCurrency(round.UserId)
		if readErr != nil {
			return nil, readErr
		}
		return &services.Luck5Resp{Code: services.ErrorCode_NO_ENOUGH_POOL_MONEY, CurrencyCent: balance, Error: "poolInsufficient"}, nil
	}
	code, err := RedisIns().commitLuck5Funds(ctx, agent, round, true)
	if err != nil {
		return &services.Luck5Resp{Code: code, Error: "settleFundsFailed"}, err
	}
	if err = saveRecord(round); err != nil {
		return &services.Luck5Resp{Code: services.ErrorCode_SYSTEM_ERROR, CurrencyCent: round.Balance, Error: "recordSaveFailed"}, err
	}
	if err = RedisIns().cli.Del(ctx, luck5Key(round.RoundId)).Err(); err != nil {
		return &services.Luck5Resp{Code: services.ErrorCode_SYSTEM_ERROR, CurrencyCent: round.Balance, Error: "roundContextClearFailed"}, err
	}
	return &services.Luck5Resp{Code: services.ErrorCode_OK, Ret: true, CurrencyCent: round.Balance}, nil
}
