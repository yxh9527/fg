package dao

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestLuck5DealPoolDeductsRevenueBeforeReserveCheck(t *testing.T) {
	pool := decimal.NewFromInt(5)
	bet := decimal.NewFromInt(10)
	revenue := decimal.NewFromInt(1)
	reserve := decimal.NewFromInt(15)
	available := luck5DealPoolAfter(pool, bet, revenue)
	if !available.Equal(decimal.NewFromInt(14)) {
		t.Fatalf("available pool = %s, want 14", available)
	}
	if available.GreaterThanOrEqual(reserve) {
		t.Fatal("pool check ignored tax deducted from the bet")
	}
	if !luck5DealPoolAfter(decimal.Zero, decimal.NewFromInt(10), decimal.Zero).GreaterThanOrEqual(decimal.NewFromInt(10)) {
		t.Fatal("zero tax should preserve the full bet contribution")
	}
}

func TestLuck5SettleAllowsAwardAboveReserveWhenPoolCanCoverDifference(t *testing.T) {
	if !luck5CanSettle(decimal.NewFromInt(30), decimal.NewFromInt(20), decimal.NewFromInt(50)) {
		t.Fatal("pool plus reserve should cover the full award")
	}
	if luck5CanSettle(decimal.NewFromInt(29), decimal.NewFromInt(20), decimal.NewFromInt(50)) {
		t.Fatal("accepted award above available funds")
	}
	if !luck5CanSettle(decimal.Zero, decimal.NewFromInt(20), decimal.NewFromInt(20)) {
		t.Fatal("award equal to reserve should not require available pool")
	}
}
