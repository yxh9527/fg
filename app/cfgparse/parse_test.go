package cfgparse

import (
	"sync"
	"testing"

	"app/config"

	"github.com/shopspring/decimal"
)

func TestParsePoolConfigUsesKeySymbol(t *testing.T) {
	config.CfgIns = &config.Configs{
		Lock: &sync.RWMutex{},
		Pool: &config.PoolMgr{
			Default: make(map[string]*config.Pool),
			Agent:   make(map[string]*config.Pool),
		},
		Award:       &config.AwardMgr{Data: make(map[string]*config.AwardConfig)},
		Coefficient: config.DefaultControlCoefficientConfig(),
	}

	raw := `{"symbol":"other","name":"test","nameZH":"测试","gameId":3037,"pool":{"1":{"min":300,"normal":900,"max":3000,"normalRate":0.99,"minRate":0.5,"maxRate":0.8,"control":1,"revenue":0.03,"base":100}}}`
	Parse("/fg/config/pool/sjddj", raw)
	got := config.CfgIns.GetPoolCfg(1, "sjddj")
	if got == nil || got.Pool == nil || got.Pool[1] == nil {
		t.Fatalf("prefixed pool key should load by key symbol")
	}
	if !got.Pool[1].Base.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("base want 100 got %s", got.Pool[1].Base)
	}

	Parse("/config/pool/sjddj", `{"symbol":"other","gameId":3037,"pool":{"1":{"base":200,"revenue":0.03,"min":1,"max":2,"normal":1,"minRate":1,"maxRate":1,"normalRate":1,"control":1}}}`)
	got = config.CfgIns.GetPoolCfg(1, "sjddj")
	if got == nil || got.Pool[1] == nil || !got.Pool[1].Base.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("unprefixed /config/pool key should be skipped")
	}
}

func TestParseControlCoefficientPrefixOnly(t *testing.T) {
	config.CfgIns = &config.Configs{
		Lock:        &sync.RWMutex{},
		Pool:        &config.PoolMgr{Default: map[string]*config.Pool{}, Agent: map[string]*config.Pool{}},
		Award:       &config.AwardMgr{Data: map[string]*config.AwardConfig{}},
		Coefficient: config.DefaultControlCoefficientConfig(),
	}
	Parse("/fg/config/controlCoefficient", `{"single_rounds":100,"bet":[{"min":0,"max":1,"coefficient":0.2}],"single":[{"min":1,"coefficient":0.5}]}`)
	got := config.CfgIns.GetControlCoefficient(decimal.RequireFromString("0.5"), decimal.NewFromInt(10), decimal.NewFromInt(1), decimal.NewFromInt(1), false)
	if !got.Equal(decimal.RequireFromString("0.2")) {
		t.Fatalf("prefixed controlCoefficient not applied, got %s", got)
	}
	Parse("/config/controlCoefficient", `{"single_rounds":100,"bet":[{"min":0,"max":1,"coefficient":0.3}],"single":[{"min":1,"coefficient":0.5}]}`)
	got = config.CfgIns.GetControlCoefficient(decimal.RequireFromString("0.5"), decimal.NewFromInt(10), decimal.NewFromInt(1), decimal.NewFromInt(1), false)
	if !got.Equal(decimal.RequireFromString("0.2")) {
		t.Fatalf("unprefixed controlCoefficient should be skipped, got %s", got)
	}
}
