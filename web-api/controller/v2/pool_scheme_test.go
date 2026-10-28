package v2

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSchemeWritesRejectInvalidScopeAndImplicitAllBeforeDatabase(t *testing.T) {
	for _, body := range []string{`{"groups":[]}`, `{"scope":"all","webId":7,"groups":[]}`, `{"scope":"site","webId":0,"groups":[]}`} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest("POST", "/pool-schemes/publish", bytes.NewBufferString(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		PoolSchemePublish(ctx)
		var response struct{ Code int }
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Code != 400 {
			t.Fatal("implicit full-scope write accepted", recorder.Body.String())
		}
	}
}

func TestFormalAgentsCannotBypassSchemeVersionCoordinationThroughLegacyWrites(t *testing.T) {
	for _, handler := range []gin.HandlerFunc{PoolGroupCreate, PoolGroupUpdate, PoolGroupDelete} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest("POST", "/pool-groups", bytes.NewBufferString(`{"agentId":7,"requestId":"legacy","name":"group","symbols":["a"]}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler(ctx)
		var response struct {
			Code int
			Msg  string
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Code != 400 || !strings.Contains(response.Msg, "专属配置") {
			t.Fatal(recorder.Body.String())
		}
	}
}
