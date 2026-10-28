package dao

import (
	"app/poollog"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"strconv"
)

// PoolSession pins an unfinished chain to its original pool. Amount is in request currency units.
type PoolSession struct {
	Kind       string `json:"kind"`
	Symbol     string `json:"symbol"`
	PoolSymbol string `json:"poolSymbol"`
	Version    int64  `json:"version"`
	RoundID    string `json:"roundId"`
	UserID     uint32 `json:"userId"`
	Amount     string `json:"amount"`
}

func PoolSessionKey(kind string, userID, gameID uint32, currency, round string) (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.String("kind", kind))
	poolLogFields = append(poolLogFields, zap.Uint32("userId", userID))
	poolLogFields = append(poolLogFields, zap.Uint32("gameId", gameID))
	poolLogFields = append(poolLogFields, zap.String("currencyType", currency))
	poolLogFields = append(poolLogFields, zap.String("roundId", round))
	poolLogTrace := poollog.BeginDebug("dao.PoolSessionKey", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	raw, _ := json.Marshal([]string{kind, strconv.FormatUint(uint64(userID), 10), strconv.FormatUint(uint64(gameID), 10), currency, round})
	sum := sha256.Sum256(raw)
	return kind + ":" + hex.EncodeToString(sum[:])
}
func Luck5SessionKey(userID, gameID uint32, currency, round string) (poolLogResult0 string) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, zap.Uint32("userId", userID))
	poolLogFields = append(poolLogFields, zap.Uint32("gameId", gameID))
	poolLogFields = append(poolLogFields, zap.String("currencyType", currency))
	poolLogFields = append(poolLogFields, zap.String("roundId", round))
	poolLogTrace := poollog.BeginDebug("dao.Luck5SessionKey", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	return PoolSessionKey("luck5", userID, gameID, currency, round)
}
func RemainingPoolSession(current, added, used string) (poolLogResult0 string, poolLogResult1 bool) {
	poolLogFields := []zap.Field{}
	poolLogTrace := poollog.BeginDebug("dao.RemainingPoolSession", poolLogFields...)
	defer poolLogTrace.Finish(nil)
	a, ea := decimal.NewFromString(current)
	b, eb := decimal.NewFromString(added)
	c, ec := decimal.NewFromString(used)
	if ea != nil || eb != nil || ec != nil {
		poollog.Event(zap.ErrorLevel, "dao.RemainingPoolSession", "amount_parse_failed", "活动链金额解析失败，保留活动状态等待核实", zap.String("currentAmount", current), zap.String("addedAmount", added), zap.String("usedAmount", used), zap.NamedError("currentError", ea), zap.NamedError("addedError", eb), zap.NamedError("usedError", ec))
		return current, true
	}
	next := a.Add(b).Sub(c)
	return next.String(), next.GreaterThan(decimal.Zero)
}
