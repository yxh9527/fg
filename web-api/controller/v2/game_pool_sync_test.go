package v2

import (
	"app/config"
	"app/entity"
	"app/esindex"
	"app/tables/manager"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSyncAllPoolUsesGameClass(t *testing.T) {
	for _, sourceClass := range []int{1, 2} {
		t.Run(fmt.Sprintf("class%d", sourceClass), func(t *testing.T) {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
			if err != nil {
				t.Fatal(err)
			}
			sqlDB, err := db.DB()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = sqlDB.Close() })
			if err := db.AutoMigrate(&manager.Game{}); err != nil {
				t.Fatal(err)
			}
			games := []manager.Game{
				{Number: 2001, ConfName: "bxjg", Name: "bxjg", NameZH: "变形金刚", GameType: 1, GameClass: 1},
				{Number: 7008, ConfName: "gw", Name: "gw", NameZH: "黄金大转轮", GameType: 2, GameClass: 1},
				{Number: 7000, ConfName: "frt", Name: "frt", NameZH: "水果机", GameType: 2, GameClass: 2},
				{Number: 7001, ConfName: "bbs", Name: "bbs", NameZH: "飞禽走兽", GameType: 2, GameClass: 2},
				{Number: 7002, ConfName: "bzb", Name: "bzb", NameZH: "奔驰宝马", GameType: 2, GameClass: 2},
				{Number: 7026, ConfName: "bcbm", Name: "bcbm", NameZH: "奔驰宝马2", GameType: 2, GameClass: 2},
				{Number: 9999, ConfName: "unclassified", GameType: 2, GameClass: 0},
			}
			if err := db.Create(&games).Error; err != nil {
				t.Fatal(err)
			}
			server := miniredis.RunT(t)
			client := redis.NewClient(&redis.Options{Addr: server.Addr()})
			t.Cleanup(func() { _ = client.Close() })
			ctx := context.Background()
			for _, game := range games {
				if err := client.Set(ctx, esindex.ConfigKey("pool", game.ConfName), "original-"+game.ConfName, time.Hour).Err(); err != nil {
					t.Fatal(err)
				}
			}
			// An agent override is outside the default-pool sync scope.
			agentKey := esindex.AgentKey("7", "pool", "frt")
			if err := client.Set(ctx, agentKey, "agent-override", 0).Err(); err != nil {
				t.Fatal(err)
			}
			sub := client.Subscribe(ctx, "message")
			t.Cleanup(func() { _ = sub.Close() })
			if _, err := sub.Receive(ctx); err != nil {
				t.Fatal(err)
			}
			sourceID := int64(7008)
			if sourceClass == 2 {
				sourceID = 7000
			}
			pool := &config.Pool{GameId: sourceID, Pool: map[int32]*config.PoolItem{1: {Min: decimal.NewFromInt(123), Base: decimal.NewFromInt(456)}}}
			if err := syncAllPool(ctx, db, client, pool); err != nil {
				t.Fatal(err)
			}
			expected := map[string]string{}
			for _, game := range games {
				key := esindex.ConfigKey("pool", game.ConfName)
				value, err := client.Get(ctx, key).Result()
				if err != nil {
					t.Fatal(err)
				}
				if game.GameClass != sourceClass {
					if value != "original-"+game.ConfName || server.TTL(key) != time.Hour {
						t.Fatalf("different-class game %s changed", game.ConfName)
					}
					continue
				}
				var got config.Pool
				if err := json.Unmarshal([]byte(value), &got); err != nil {
					t.Fatal(err)
				}
				if got.GameId != int64(game.Number) || got.Symbol != game.ConfName || got.Name != game.Name || got.NameZH != game.NameZH {
					t.Fatalf("target identity changed: %+v", got)
				}
				if got.Pool[1] == nil || !got.Pool[1].Min.Equal(decimal.NewFromInt(123)) || !got.Pool[1].Base.Equal(decimal.NewFromInt(456)) {
					t.Fatalf("same-class game %s did not receive pool values", game.ConfName)
				}
				expected[key] = value
			}
			for range expected {
				receiveCtx, cancel := context.WithTimeout(ctx, time.Second)
				message, err := sub.ReceiveMessage(receiveCtx)
				cancel()
				if err != nil {
					t.Fatal(err)
				}
				var envelope entity.Msg
				if err := json.Unmarshal([]byte(message.Payload), &envelope); err != nil {
					t.Fatal(err)
				}
				var change entity.ConfigMsg
				if err := json.Unmarshal([]byte(envelope.Data), &change); err != nil {
					t.Fatal(err)
				}
				if want, ok := expected[change.Key]; envelope.Event != "config" || !ok || change.Data != want {
					t.Fatalf("unexpected config broadcast: %+v", change)
				}
			}
			if client.Get(ctx, agentKey).Val() != "agent-override" {
				t.Fatal("agent override changed")
			}
			for _, invalidID := range []int64{8888, 9999} {
				before := server.Dump()
				if err := syncAllPool(ctx, db, client, &config.Pool{GameId: invalidID}); err == nil {
					t.Fatalf("invalid source %d accepted", invalidID)
				}
				if after := server.Dump(); after != before {
					t.Fatalf("invalid source %d changed Redis", invalidID)
				}
			}
		})
	}
}
