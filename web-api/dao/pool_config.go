package dao

import (
	"app/tables/manager"
	"context"

	"go.uber.org/zap"
)

func SyncPoolConfigsToRedis() error {
	var poolConfigs []manager.PoolConfig
	if err := Mysql().Manager.Find(&poolConfigs).Error; err != nil {
		return err
	}

	pipe := RedisIns().Client.Pipeline()
	for _, item := range poolConfigs {
		if item.Key == "" {
			continue
		}
		pipe.Set(context.Background(), item.Key, item.Value, 0)
	}

	if _, err := pipe.Exec(context.Background()); err != nil {
		return err
	}

	zap.L().Info("pool_config loaded to redis", zap.Int("count", len(poolConfigs)))
	return nil
}
