package crontab

import (
	"app/entity"
	"app/entity/view"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type failResetPublish struct{}

func (failResetPublish) DialHook(next redis.DialHook) redis.DialHook { return next }
func (failResetPublish) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}
func (failResetPublish) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if cmd.Name() == "publish" {
			return errors.New("publish unavailable")
		}
		return next(ctx, cmd)
	}
}

func TestManualResetDispatchesImmediatelyDespiteFutureSchedule(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer rdb.Close()
	ctx := context.Background()
	now := time.Unix(1793265000, 0)
	schedule := view.ResetPoolConfig{Interval: 7, ResetTime: now.Unix() + 86400}
	raw, _ := json.Marshal(schedule)
	if err := rdb.Set(ctx, "reset_pool_config", raw, 0).Err(); err != nil {
		t.Fatal(err)
	}
	sub := rdb.Subscribe(ctx, "message")
	defer sub.Close()
	if _, err := sub.Receive(ctx); err != nil {
		t.Fatal(err)
	}
	batch, err := dispatchPoolReset(ctx, rdb, true, now)
	if err != nil || batch == "" {
		t.Fatal("manual reset still waits for cron", batch, err)
	}
	receiveCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	message, err := sub.ReceiveMessage(receiveCtx)
	if err != nil {
		t.Fatal(err)
	}
	var envelope entity.Msg
	if err = json.Unmarshal([]byte(message.Payload), &envelope); err != nil {
		t.Fatal(err)
	}
	var reset entity.ResetPool
	if err = json.Unmarshal([]byte(envelope.Data), &reset); err != nil {
		t.Fatal(err)
	}
	if reset.BatchID != batch || reset.Publisher != "web-api.PoolResetNow" {
		t.Fatal(reset)
	}
	updated, err := rdb.Get(ctx, "reset_pool_config").Bytes()
	if err != nil {
		t.Fatal(err)
	}
	var next view.ResetPoolConfig
	if err = json.Unmarshal(updated, &next); err != nil {
		t.Fatal(err)
	}
	if next.Now || next.ResetTime != now.Unix()+7*86400 {
		t.Fatal("next reset schedule was not maintained", next)
	}
}

func TestScheduledResetDoesNotDispatchBeforeDueOrAdvanceFailedPublish(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	defer rdb.Close()
	ctx := context.Background()
	now := time.Unix(1793265000, 0)
	schedule := view.ResetPoolConfig{Interval: 7, ResetTime: now.Unix() + 1000}
	raw, _ := json.Marshal(schedule)
	if err := rdb.Set(ctx, "reset_pool_config", raw, 0).Err(); err != nil {
		t.Fatal(err)
	}
	batch, err := dispatchPoolReset(ctx, rdb, false, now)
	if err != nil || batch != "" {
		t.Fatal("scheduled reset ran before due", batch, err)
	}
	rdb.AddHook(failResetPublish{})
	if _, err = dispatchPoolReset(ctx, rdb, true, now); err == nil || err.Error() != "publish unavailable" {
		t.Fatal("publication failure was not reported", err)
	}
	got, err := rdb.Get(ctx, "reset_pool_config").Result()
	if err != nil || got != string(raw) {
		t.Fatal("failed publication changed schedule", got, err)
	}
}
