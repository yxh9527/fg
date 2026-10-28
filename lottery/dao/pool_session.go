package dao

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/shopspring/decimal"
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

func PoolSessionKey(kind string, userID, gameID uint32, currency, round string) string {
	raw, _ := json.Marshal([]string{kind, strconv.FormatUint(uint64(userID), 10), strconv.FormatUint(uint64(gameID), 10), currency, round})
	sum := sha256.Sum256(raw)
	return kind + ":" + hex.EncodeToString(sum[:])
}
func Luck5SessionKey(userID, gameID uint32, currency, round string) string {
	return PoolSessionKey("luck5", userID, gameID, currency, round)
}
func RemainingPoolSession(current, added, used string) (string, bool) {
	a, ea := decimal.NewFromString(current)
	b, eb := decimal.NewFromString(added)
	c, ec := decimal.NewFromString(used)
	if ea != nil || eb != nil || ec != nil {
		return current, true
	}
	next := a.Add(b).Sub(c)
	return next.String(), next.GreaterThan(decimal.Zero)
}
