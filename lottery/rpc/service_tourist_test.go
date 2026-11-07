package rpc

import (
	"app/config"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/shopspring/decimal"
)

func TestRecordTouristFlagSurvivesConversionAndSave(t *testing.T) {
	previous := config.CfgIns
	t.Cleanup(func() { config.CfgIns = previous })
	config.CfgIns = &config.Configs{
		Lock:     &sync.RWMutex{},
		Currency: &config.CurrencyMgr{Data: map[string]decimal.Decimal{"CNY": decimal.NewFromInt(1)}},
		Pool: &config.PoolMgr{
			Default: map[string]*config.Pool{"game": {Symbol: "game", GameId: 1, Pool: map[int32]*config.PoolItem{1: {Revenue: decimal.RequireFromString("0.1")}}}},
			Agent:   map[string]*config.Pool{},
		},
	}
	service := &LotteryService{recordsCache: &RecordCacheMgr{lock: &sync.RWMutex{}, records: map[string]*RecordItem{}}}
	for _, flag := range []int32{0, 1} {
		t.Run(fmt.Sprintf("isTourist=%d", flag), func(t *testing.T) {
			record := ConvertRecord(7, 1, fmt.Sprintf("round-%d", flag), "CNY", "game", "player", `{"isTourist":0}`, decimal.NewFromInt(105), 23, 10, 15, flag)
			service.SaveRecord(record)
			cached := service.recordsCache.records[record.Hash]
			if cached == nil || cached.record.IsTourist != flag {
				t.Fatalf("saved record lost verified player flag: %+v", cached)
			}
			data, err := json.Marshal(cached.record)
			if err != nil {
				t.Fatal(err)
			}
			var stored map[string]any
			if err := json.Unmarshal(data, &stored); err != nil {
				t.Fatal(err)
			}
			if stored["isTourist"] != float64(flag) || stored["bet"] != float64(10) || stored["win"] != float64(15) || stored["balance"] != float64(105) {
				t.Fatalf("incorrect persisted flag or amounts: %s", data)
			}
		})
	}
}
