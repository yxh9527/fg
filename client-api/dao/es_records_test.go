package dao

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"app/esindex"

	"github.com/olivere/elastic/v7"
)

// TestListRecordsSummaryIgnoresPagination 验证总计使用同一筛选下的所有记录，翻页和空页不改变总计。
func TestListRecordsSummaryIgnoresPagination(t *testing.T) {
	query := RecordListQuery{
		UserId: 279, GameId: 2018, Symbol: "game", Currency: "CNY",
		StartMs: 1795233600000, EndMs: 1795256669000, Size: 1,
	}
	wantQuery := fmt.Sprintf(`{"bool":{"must":[
		{"term":{"userId":279}}, {"term":{"complete":true}},
		{"bool":{"minimum_should_match":"1","should":[{"term":{"gameId":2018}},{"term":{"symbol":"game"}}]}},
		{"match_phrase":{"currency":{"query":"CNY"}}},
		{"range":{"playedDate":{"from":%d,"to":%d,"include_lower":true,"include_upper":true}}}
	]}}`, query.StartMs, query.EndMs)
	var expectedQuery interface{}
	if err := json.Unmarshal([]byte(wantQuery), &expectedQuery); err != nil {
		t.Fatal(err)
	}
	var requestedPage int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+esindex.Settlement()+"/_search" {
			t.Errorf("unexpected ES path: %s", r.URL.Path)
		}
		var request struct {
			From         int                    `json:"from"`
			Size         int                    `json:"size"`
			Query        interface{}            `json:"query"`
			Aggregations map[string]interface{} `json:"aggregations"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(request.Query, expectedQuery) {
			t.Errorf("scope differs: got %#v, want %#v", request.Query, expectedQuery)
		}
		if request.From != requestedPage-1 || request.Size != 1 {
			t.Errorf("unexpected pagination: from=%d size=%d", request.From, request.Size)
		}
		for name, field := range map[string]string{"totalBet": "bet", "totalWin": "win"} {
			want := map[string]interface{}{"sum": map[string]interface{}{"field": field}}
			if !reflect.DeepEqual(request.Aggregations[name], want) {
				t.Errorf("missing unpaginated %s sum: %#v", field, request.Aggregations)
			}
		}
		hits := []interface{}{}
		if requestedPage <= 3 {
			hits = append(hits, map[string]interface{}{"_source": RecordItem{
				UserId: 279, GameId: 2018, Bet: float64(requestedPage) / 10, Win: 0.15,
			}})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"hits": map[string]interface{}{"total": map[string]interface{}{"value": 3, "relation": "eq"}, "hits": hits},
			"aggregations": map[string]interface{}{
				"totalBet": map[string]float64{"value": 0.6},
				"totalWin": map[string]float64{"value": 0.45},
			},
		})
	}))
	defer server.Close()
	client, err := elastic.NewClient(elastic.SetURL(server.URL), elastic.SetSniff(false), elastic.SetHealthcheck(false))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Stop()
	dao := &ESDao{es: client}
	for _, page := range []int{1, 2, 4} {
		requestedPage, query.Page = page, page
		result, err := dao.ListRecords(query)
		if err != nil {
			t.Fatal(err)
		}
		if result.Total != 3 || result.Summary != (RecordSummary{Bet: 0.6, Win: 0.45}) {
			t.Fatalf("page %d changed totals: %+v", page, result)
		}
		if page <= 3 && (len(result.List) != 1 || result.List[0].Bet != float64(page)/10) {
			t.Fatalf("wrong page %d records: %+v", page, result.List)
		}
		if page == 4 && len(result.List) != 0 {
			t.Fatalf("out-of-range page returned records: %+v", result.List)
		}
	}
}

// TestListRecordsEmptyAndMissingSummary 验证无匹配记录返回零，缺少聚合不能伪装成正确总计。
func TestListRecordsEmptyAndMissingSummary(t *testing.T) {
	for _, tc := range []struct {
		name         string
		aggregations string
		wantError    bool
	}{
		{"empty", `{"totalBet":{"value":0},"totalWin":{"value":0}}`, false},
		{"missing", `{}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintf(w, `{"hits":{"total":{"value":0,"relation":"eq"},"hits":[]},"aggregations":%s}`, tc.aggregations)
			}))
			defer server.Close()
			client, err := elastic.NewClient(elastic.SetURL(server.URL), elastic.SetSniff(false), elastic.SetHealthcheck(false))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Stop()
			result, err := (&ESDao{es: client}).ListRecords(RecordListQuery{UserId: 279, GameId: 2018})
			if tc.wantError {
				if err == nil {
					t.Fatal("accepted missing summary")
				}
				return
			}
			if err != nil || result == nil || result.Total != 0 || len(result.List) != 0 || result.Summary != (RecordSummary{}) {
				t.Fatalf("unexpected empty result: %+v, %v", result, err)
			}
		})
	}
}
