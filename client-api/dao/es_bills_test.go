package dao

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"app/esindex"

	"github.com/olivere/elastic/v7"
)

// TestListBillsSettlementTypes 验证两种完局流水都被 ES 查询和读取过滤保留，普通流水查询不受影响。
func TestListBillsSettlementTypes(t *testing.T) {
	for _, onlySettlement := range []bool{true, false} {
		name := "all"
		if onlySettlement {
			name = "settlement"
		}
		t.Run(name, func(t *testing.T) {
			wantTypes := []string{"返奖", "结算"}
			if !onlySettlement {
				wantTypes = append(wantTypes, "下注", "回退")
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path != "/"+esindex.FlowingWater()+"/_search" {
					t.Errorf("unexpected ES path: %s", r.URL.Path)
					_, _ = w.Write([]byte(`{"count":2}`))
					return
				}
				var request struct {
					Query struct {
						Bool struct {
							Must []json.RawMessage `json:"must"`
						} `json:"bool"`
					} `json:"query"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				wantCount := 2
				if onlySettlement {
					wantCount++
				}
				if len(request.Query.Bool.Must) != wantCount {
					t.Errorf("unexpected filter count: %d", len(request.Query.Bool.Must))
				} else if onlySettlement {
					var got, want interface{}
					_ = json.Unmarshal(request.Query.Bool.Must[2], &got)
					_ = json.Unmarshal([]byte(`{"bool":{"minimum_should_match":"1","should":[
						{"term":{"desc.keyword":"返奖"}},
						{"term":{"desc":"返奖"}},
						{"match_phrase":{"desc":{"query":"返奖"}}},
						{"term":{"desc.keyword":"结算"}},
						{"term":{"desc":"结算"}},
						{"match_phrase":{"desc":{"query":"结算"}}}
					]}}`), &want)
					if !reflect.DeepEqual(got, want) {
						t.Errorf("settlement filter differs: got %#v, want %#v", got, want)
					}
				}
				hits := make([]interface{}, 0, len(wantTypes))
				for _, desc := range wantTypes {
					hits = append(hits, map[string]interface{}{"_source": BillItem{
						UserId: 7, GameId: 7018, Desc: desc, Bet: -10, Award: 12,
					}})
				}
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"hits": map[string]interface{}{
						"total": map[string]interface{}{"value": len(hits), "relation": "eq"},
						"hits":  hits,
					},
				})
			}))
			defer server.Close()
			client, err := elastic.NewClient(elastic.SetURL(server.URL), elastic.SetSniff(false), elastic.SetHealthcheck(false))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Stop()
			items, total, err := (&ESDao{es: client}).ListBills(BillListQuery{
				UserId: 7, GameId: 7018, OnlySettlement: onlySettlement,
			})
			if err != nil {
				t.Fatal(err)
			}
			var gotTypes []string
			for _, item := range items {
				gotTypes = append(gotTypes, item.Desc)
				if item.Bet != -10 || item.Award != 12 {
					t.Errorf("amounts changed: %+v", item)
				}
			}
			if !reflect.DeepEqual(gotTypes, wantTypes) || total != int64(len(wantTypes)) {
				t.Fatalf("got types=%v total=%d, want types=%v", gotTypes, total, wantTypes)
			}
		})
	}
}
