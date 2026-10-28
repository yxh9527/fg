package v2

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPoolGroupAgentAllowsVirtualZeroWithoutAgentTable(t *testing.T) {
	for _, query := range []string{"agentId=0", "agentId=0&webId=123"} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest("GET", "/pool-groups?"+query, nil)
		id, err := poolGroupAgent(ctx)
		if err != nil || id != 0 {
			t.Fatal("virtual test agent zero was rejected", id, err)
		}
	}
}
func TestPoolGroupAgentRejectsMissingAndNegativeSelection(t *testing.T) {
	for _, query := range []string{"", "agentId=", "agentId=-1", "agentId=invalid"} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest("GET", "/pool-groups?"+query, nil)
		if _, err := poolGroupAgent(ctx); err == nil {
			t.Fatal("accepted invalid selection", query)
		}
	}
}
func TestPoolGroupBodyDistinguishesExplicitZeroAndNoAgent(t *testing.T) {
	cases := []struct {
		body    map[string]interface{}
		present bool
		want    int64
	}{{body: map[string]interface{}{"agentId": 0, "name": "test", "symbols": []string{"a"}}, present: true, want: 0}, {body: map[string]interface{}{"agentId": 7}, present: true, want: 7}, {body: map[string]interface{}{"name": "test"}}, {body: map[string]interface{}{"agentId": nil}}}
	for _, item := range cases {
		body, err := json.Marshal(item.body)
		if err != nil {
			t.Fatal(err)
		}
		var input poolGroupInput
		if err = json.Unmarshal(body, &input); err != nil {
			t.Fatal(err)
		}
		if (input.AgentID != nil) != item.present {
			t.Fatal("incorrect agent presence", item.body)
		}
		if item.present && *input.AgentID != item.want {
			t.Fatal(input.AgentID)
		}
	}
}
func TestPoolGroupWriteRejectsOmittedAgentBeforeDatabaseAccess(t *testing.T) {
	for _, handler := range []gin.HandlerFunc{PoolGroupCreate, PoolGroupUpdate, PoolGroupDelete} {
		for _, input := range []map[string]interface{}{{"name": "test"}, {"agentId": nil, "name": "test"}} {
			body, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest("POST", "/pool-groups", bytes.NewReader(body))
			ctx.Request.Header.Set("Content-Type", "application/json")
			handler(ctx)
			var response struct {
				Code int
				Msg  string
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Code != 400 || !strings.Contains(response.Msg, "选择代理") {
				t.Fatal("missing agent was silently treated as test zero", recorder.Body.String())
			}
		}
	}
}
