package crontab

import (
	"app/entity"
	"app/entity/view"
	"app/poollog"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
	"web-api/dao"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Manual reset publishes now; it does not create SQL reset operations.
func TriggerPoolResetNow(ctx context.Context) (string, error) { return triggerPoolReset(ctx, true) }
func triggerPoolReset(ctx context.Context, manual bool) (string, error) {
	return dispatchPoolReset(ctx, dao.RedisIns().Client, manual, time.Now())
}
func dispatchPoolReset(ctx context.Context, rdb redis.UniversalClient, manual bool, now time.Time) (string, error) {
	raw, err := rdb.Get(ctx, "reset_pool_config").Bytes()
	if err != nil {
		return "", err
	}
	var schedule view.ResetPoolConfig
	if err = json.Unmarshal(raw, &schedule); err != nil {
		return "", err
	}
	if schedule.ResetTime <= 0 || schedule.Interval <= 0 || schedule.Interval > 31 {
		return "", errors.New("水池重置排期或周期无效")
	}
	if !manual && !schedule.Now && schedule.ResetTime > now.Unix() {
		return "", nil
	}
	batchID := fmt.Sprintf("%d-%t", schedule.ResetTime, manual || schedule.Now)
	host, _ := os.Hostname()
	instance := fmt.Sprintf("%s:%d", host, os.Getpid())
	publisher := "web-api.crontab.clearPool"
	if manual {
		publisher = "web-api.PoolResetNow"
	}
	message, err := entity.NewResetPoolMessage(entity.ResetPool{BatchID: batchID, Publisher: publisher, PublisherInstance: instance, PublishedAt: now.Unix()})
	if err != nil {
		return "", err
	}
	poollog.Event(zap.InfoLevel, "crontab.clearPool", "reset_triggered", "触发直接清空水池批次", zap.String("batchId", batchID), zap.Bool("manual", manual || schedule.Now))
	if err = rdb.Publish(ctx, "message", message).Err(); err != nil {
		poollog.Event(zap.ErrorLevel, "crontab.clearPool", "publish_failed", "水池重置消息发送失败，未推进排期", zap.String("batchId", batchID), zap.Error(err))
		return "", err
	} else {
		poollog.Event(zap.InfoLevel, "crontab.clearPool", "reset_published", "直接重置已通知开奖服务执行", zap.String("batchId", batchID), zap.String("publisher", publisher), zap.String("publisherInstance", instance))
	}
	schedule.Now = false
	schedule.ResetTime = now.Unix() + int64(schedule.Interval)*24*60*60
	updated, _ := json.Marshal(schedule)
	if err = rdb.Set(ctx, "reset_pool_config", updated, 0).Err(); err != nil {
		poollog.Event(zap.ErrorLevel, "crontab.clearPool", "schedule_write_failed", "重置消息已发送，但下一次排期保存失败", zap.String("batchId", batchID), zap.Error(err))
	} else {
		poollog.Event(zap.InfoLevel, "crontab.clearPool", "next_schedule_saved", "下一次排期已保存，开奖服务收到消息后直接清内存及四类累计", zap.String("batchId", batchID), zap.Int64("nextResetTime", schedule.ResetTime))
	}
	return batchID, nil
}
