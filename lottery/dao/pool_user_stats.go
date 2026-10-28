package dao

import (
	"app/poollog"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"strconv"
)

// A missing cumulative counter is initially zero; a transport/type/identity error must not reset statistics.
func (rd *RedisDao) GetUserStatDataChecked(id uint32) (poolLogResult0 decimal.Decimal, poolLogResult1 decimal.Decimal, poolLogResult2 decimal.Decimal, poolLogResult3 string, poolLogResult4 int32, poolLogResult5 error) {
	poolLogFields := []zap.Field{}
	poolLogFields = append(poolLogFields, poollog.MetadataFields(rd)...)
	poolLogFields = append(poolLogFields, zap.Uint32("id", id))
	poolLogTrace := poollog.Begin("dao.RedisDao.GetUserStatDataChecked", zap.DebugLevel, poolLogFields...)
	defer poolLogTrace.Finish(&poolLogResult5)
	ctx := context.Background()
	member := strconv.FormatUint(uint64(id), 10)
	pipe := rd.cli.Pipeline()
	commands := []*redis.FloatCmd{pipe.ZScore(ctx, "userTotalEffBet", member), pipe.ZScore(ctx, "userTotalProfLoss", member), pipe.ZScore(ctx, "userBetCount", member)}
	account := pipe.HGet(ctx, "player_"+member, "account")
	tourist := pipe.HGet(ctx, "player_"+member, "isTourist")
	if _, err := pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		if err != nil {
			poollog.Event(zap.ErrorLevel, "dao.RedisDao.GetUserStatDataChecked", "pipe.Exec", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		}
		return decimal.Zero, decimal.Zero, decimal.Zero, "", 0, err
	}
	values := make([]decimal.Decimal, 3)
	for i, command := range commands {
		value, err := command.Result()
		if errors.Is(err, redis.Nil) {
			values[i] = decimal.Zero
			continue
		}
		if err != nil {
			poollog.Event(zap.ErrorLevel, "dao.RedisDao.GetUserStatDataChecked", "command.Result", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
			return decimal.Zero, decimal.Zero, decimal.Zero, "", 0, err
		}
		values[i] = decimal.NewFromFloat(value)
	}
	name, err := account.Result()
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.GetUserStatDataChecked", "account.Result", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return decimal.Zero, decimal.Zero, decimal.Zero, "", 0, fmt.Errorf("读取玩家身份失败: %w", err)
	}
	raw, err := tourist.Result()
	if err != nil {
		poollog.Event(zap.ErrorLevel, "dao.RedisDao.GetUserStatDataChecked", "tourist.Result", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		return decimal.Zero, decimal.Zero, decimal.Zero, "", 0, err
	}
	kind, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || (kind != 0 && kind != 1) {
		if err != nil {
			poollog.Event(zap.ErrorLevel, "dao.RedisDao.GetUserStatDataChecked", "strconv.ParseInt", "水池步骤失败", append(poolLogFields, zap.Error(err))...)
		}
		return decimal.Zero, decimal.Zero, decimal.Zero, "", 0, errors.New("玩家游客标记无效")
	}
	return values[0], values[1], values[2], name, int32(kind), nil
}
