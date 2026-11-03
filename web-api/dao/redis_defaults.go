package dao

import (
	"app/esindex"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var redisConfigsInitializedKey = esindex.Name("redis_configs_initialized")

// The script serializes concurrent starts and only marks a completed import.
// SET NX preserves configuration already present in an existing installation.
var initRedisConfigs = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 1 then
    return -1
end
local inserted = 0
for i = 2, #KEYS do
    if redis.call('SET', KEYS[i], ARGV[i - 1], 'NX') then
        inserted = inserted + 1
    end
end
redis.call('SET', KEYS[1], '1')
return inserted
`)

// InitRedisConfigs imports the built-in /fg configuration snapshot once per Redis DB.
func InitRedisConfigs(ctx context.Context) error {
	inserted, err := initializeRedisConfigs(ctx, RedisIns().Client)
	if err != nil {
		return err
	}
	if inserted < 0 {
		zap.L().Info("redis configs already initialized, skipping")
	} else {
		zap.L().Info("redis configs initialized", zap.Int("inserted", inserted))
	}
	return nil
}

func initializeRedisConfigs(ctx context.Context, client redis.UniversalClient) (int, error) {
	defaults := defaultRedisConfigs
	if len(defaults) == 0 {
		return 0, fmt.Errorf("default Redis configs are empty")
	}

	keys := make([]string, 0, len(defaults))
	for key := range defaults {
		if !strings.HasPrefix(key, esindex.PathRoot()) {
			return 0, fmt.Errorf("default Redis config key %q has an invalid prefix", key)
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		values = append(values, defaults[key])
	}
	keys = append([]string{redisConfigsInitializedKey}, keys...)
	inserted, err := initRedisConfigs.Run(ctx, client, keys, values...).Int()
	if err != nil {
		return 0, fmt.Errorf("initialize Redis configs: %w", err)
	}
	return inserted, nil
}
