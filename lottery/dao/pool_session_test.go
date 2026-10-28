package dao

import "testing"

func TestPoolSessionAmountsDrainWithoutRounding(t *testing.T) {
	amount, open := RemainingPoolSession("0", "10.1234", "0")
	if amount != "10.1234" || !open {
		t.Fatal(amount, open)
	}
	amount, open = RemainingPoolSession(amount, "0", "3.0001")
	if amount != "7.1233" || !open {
		t.Fatal(amount, open)
	}
	amount, open = RemainingPoolSession(amount, "0", "7.1233")
	if amount != "0" || open {
		t.Fatal(amount, open)
	}
}
func TestPoolSessionIdentityIncludesCurrencyAndUser(t *testing.T) {
	a := PoolSessionKey("multi", 1, 2, "CNY", "")
	if a == PoolSessionKey("multi", 2, 2, "CNY", "") || a == PoolSessionKey("multi", 1, 2, "USD", "") {
		t.Fatal("session identities collide")
	}
}
