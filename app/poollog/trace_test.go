package poollog_test

import (
	"app/poollog"
	"app/tables/manager"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func capture(t *testing.T, level zapcore.Level) *observer.ObservedLogs {
	t.Helper()
	core, logs := observer.New(level)
	restore := zap.ReplaceGlobals(zap.New(core))
	t.Cleanup(restore)
	return logs
}

func TestFailureStillLogsContextWhenDebugIsDisabled(t *testing.T) {
	logs := capture(t, zap.InfoLevel)
	trace := poollog.Begin("poolgroup.Store.Load", zap.DebugLevel, zap.Int64("agentId", 7), zap.String("requestId", "r1"))
	failure := errors.New("redis unavailable")
	trace.Finish(&failure)
	events := logs.FilterLevelExact(zap.ErrorLevel).All()
	if len(events) != 1 {
		t.Fatalf("expected one visible error, got %v", logs.All())
	}
	fields := events[0].ContextMap()
	if fields["function"] != "poolgroup.Store.Load" || fields["agentId"] != int64(7) || fields["stage"] != "exit" || fields["error"] != failure.Error() {
		t.Fatal(fields)
	}
	if _, ok := fields["elapsedMs"]; !ok {
		t.Fatal("error lacks elapsed time")
	}
}
func TestNormalWaitingIsWarningAndPreservesErrorsIs(t *testing.T) {
	logs := capture(t, zap.DebugLevel)
	sentinel := errors.New("waiting for unsettled round")
	waiting := poollog.Expected(sentinel)
	wrapped := fmt.Errorf("agent 7: %w", waiting)
	trace := poollog.Begin("dao.executePoolGroupOperation", zap.InfoLevel, zap.Int64("operationId", 12))
	trace.Finish(&wrapped)
	poollog.Event(zap.ErrorLevel, "dao.executePoolResetBatch", "execute", "waiting", zap.Error(wrapped))
	if !errors.Is(wrapped, sentinel) {
		t.Fatal("logging wrapper broke errors.Is")
	}
	if logs.FilterLevelExact(zap.ErrorLevel).Len() != 0 || logs.FilterLevelExact(zap.WarnLevel).Len() != 2 {
		t.Fatal(logs.All())
	}
}
func TestTraceRecordsEntryAndExit(t *testing.T) {
	logs := capture(t, zap.DebugLevel)
	trace := poollog.Begin("dao.applyPoolRuntime", zap.InfoLevel, zap.Int64("agentId", 7))
	var failure error
	trace.Finish(&failure)
	events := logs.All()
	if len(events) != 2 || events[0].ContextMap()["stage"] != "entry" || events[1].ContextMap()["stage"] != "exit" {
		t.Fatal(events)
	}
}
func TestPanicIsLoggedAndPropagatedWithoutDumpingPayload(t *testing.T) {
	logs := capture(t, zap.DebugLevel)
	payload := &struct{ Token string }{Token: "private-panic-token"}
	var recovered interface{}
	func() {
		defer func() { recovered = recover() }()
		trace := poollog.Begin("dao.test", zap.InfoLevel)
		defer trace.Finish(nil)
		panic(payload)
	}()
	if recovered != payload {
		t.Fatal("logger swallowed or changed panic")
	}
	events := logs.FilterLevelExact(zap.ErrorLevel).All()
	if len(events) != 1 || events[0].ContextMap()["stage"] != "panic" {
		t.Fatal(logs.All())
	}
	if strings.Contains(fmt.Sprint(events), payload.Token) {
		t.Fatal("panic dumped a secret object")
	}
}
func TestMetadataAllowsIDsButNeverTokensOrRawRequests(t *testing.T) {
	logs := capture(t, zap.DebugLevel)
	request := struct {
		AgentID   int64
		RequestID string
		Token     string
		Payload   string
		Config    string
		Password  string
	}{7, "r1", "private-token", "private-payload", "private-config", "private-password"}
	poollog.Event(zap.InfoLevel, "test", "metadata", "safe", poollog.MetadataFields(request)...)
	fields := logs.All()[0].ContextMap()
	if fields["agentId"] != int64(7) || fields["requestId"] != "r1" {
		t.Fatal(fields)
	}
	for _, secret := range []string{request.Token, request.Payload, request.Config, request.Password} {
		if strings.Contains(fmt.Sprint(fields), secret) {
			t.Fatal("secret present in metadata", fields)
		}
	}
	operation := manager.PoolGroupOperation{ID: 12, AgentID: 7, Kind: "reset", RequestID: "reset:batch-1"}
	poollog.Event(zap.InfoLevel, "test", "operation", "safe", poollog.MetadataFields(operation)...)
	fields = logs.All()[1].ContextMap()
	if fields["operationId"] != int64(12) || fields["batchId"] != "batch-1" {
		t.Fatal(fields)
	}
}

func TestEveryPoolBusinessMethodHasExactlyOneTrace(t *testing.T) {
	files := map[string][]string{"app/poolgroup/store.go": nil, "app/config/pool_group.go": nil, "app/tables/manager/poolGroup.go": nil, "lottery/dao/pool_group.go": nil, "lottery/dao/pool_session.go": nil, "lottery/dao/pool_user_stats.go": nil, "lottery/dao/luck5.go": nil, "lottery/rpc/pool_group.go": nil, "web-api/controller/v2/pool_group.go": nil, "lottery/dao/cache.go": {"ResetPool", "GetGame", "getPoolGame", "GetUser", "SaveAgentData", "SaveUserData", "GetAgentChanges"}, "app/config/config.go": {"GetPoolCfg", "effectivePoolLocked", "GetPoolCfgByGameId"}, "web-api/crontab/crontab.go": {"clearPool"}}
	count := 0
	for path, selected := range files {
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "..", path), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			method, ok := declaration.(*ast.FuncDecl)
			if !ok || method.Body == nil {
				continue
			}
			if selected != nil {
				match := false
				for _, name := range selected {
					match = match || method.Name.Name == name
				}
				if !match {
					continue
				}
			}
			starts, finishes := 0, 0
			ast.Inspect(method.Body, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
						if owner, ok := selector.X.(*ast.Ident); ok {
							if owner.Name == "poollog" && (selector.Sel.Name == "Begin" || selector.Sel.Name == "BeginDebug") {
								starts++
							}
							if owner.Name == "poolLogTrace" && selector.Sel.Name == "Finish" {
								finishes++
							}
						}
					}
				}
				return true
			})
			if starts != 1 || finishes != 1 {
				t.Errorf("%s.%s: entry traces=%d, deferred exits=%d", path, method.Name.Name, starts, finishes)
			}
			count++
		}
	}
	t.Logf("verified method trace coverage: %d methods", count)
}
