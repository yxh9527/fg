// Package poollog supplies consistent, correlated logs for pool grouping and reset operations.
package poollog

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Trace struct {
	logger      *zap.Logger
	level       zapcore.Level
	started     time.Time
	warnOnError bool
}

type expectedError struct{ err error }

func (e *expectedError) Error() string { return e.err.Error() }
func (e *expectedError) Unwrap() error { return e.err }
func Expected(err error) error {
	if err == nil {
		return nil
	}
	return &expectedError{err: err}
}
func isWarning(err error) bool {
	var expected *expectedError
	return errors.As(err, &expected) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
func (t *Trace) WarnOnError() {
	if t != nil {
		t.warnOnError = true
	}
}

func Begin(function string, level zapcore.Level, fields ...zap.Field) *Trace {
	logger := zap.L().With(append([]zap.Field{zap.String("module", "pool-group"), zap.String("function", function)}, fields...)...)
	trace := &Trace{logger: logger, level: level, started: time.Now()}
	if entry := logger.Check(level, "水池方法开始"); entry != nil {
		entry.Write(zap.String("stage", "entry"))
	}
	return trace
}

// BeginDebug avoids retaining traces for pure, frequently called helpers when debug is disabled.
func BeginDebug(function string, fields ...zap.Field) *Trace {
	if !zap.L().Core().Enabled(zap.DebugLevel) {
		return nil
	}
	return Begin(function, zap.DebugLevel, fields...)
}

// Finish is deferred directly so panics are logged and then propagated unchanged.
func (t *Trace) Finish(result *error) {
	if recovered := recover(); recovered != nil {
		if t != nil {
			fields := []zap.Field{zap.String("stage", "panic"), zap.String("panicType", reflect.TypeOf(recovered).String()), zap.Stack("stack"), zap.Duration("duration", time.Since(t.started))}
			switch value := recovered.(type) {
			case error:
				fields = append(fields, zap.Error(value))
			case string:
				fields = append(fields, zap.String("panicMessage", value))
			}
			t.logger.Error("水池方法发生异常", fields...)
		}
		panic(recovered)
	}
	if t == nil {
		return
	}
	fields := []zap.Field{zap.String("stage", "exit"), zap.Duration("duration", time.Since(t.started)), zap.Int64("elapsedMs", time.Since(t.started).Milliseconds())}
	if result != nil && *result != nil {
		fields = append(fields, zap.Error(*result))
		if t.warnOnError || isWarning(*result) {
			t.logger.Warn("水池方法等待、中断或受限", fields...)
		} else {
			t.logger.Error("水池方法失败", fields...)
		}
		return
	}
	if entry := t.logger.Check(t.level, "水池方法结束"); entry != nil {
		entry.Write(fields...)
	}
}

func Event(level zapcore.Level, function, stage, message string, fields ...zap.Field) {
	if level == zap.ErrorLevel {
		for _, field := range fields {
			if err, ok := field.Interface.(error); ok && isWarning(err) {
				level = zap.WarnLevel
				break
			}
		}
	}
	logger := zap.L()
	if entry := logger.Check(level, message); entry != nil {
		entry.Write(append([]zap.Field{zap.String("module", "pool-group"), zap.String("function", function), zap.String("stage", stage)}, fields...)...)
	}
}

// MetadataFields only extracts allowlisted identifiers. Never marshal a request, lease, token or config.
func MetadataFields(value interface{}) []zap.Field {
	fields := make([]zap.Field, 0, 10)
	if r, ok := value.(interface{ GetAgentId() uint32 }); ok {
		fields = append(fields, zap.Uint32("agentId", r.GetAgentId()))
	}
	if r, ok := value.(interface{ GetGameId() uint32 }); ok {
		fields = append(fields, zap.Uint32("gameId", r.GetGameId()))
	}
	if r, ok := value.(interface{ GetUserId() uint32 }); ok {
		fields = append(fields, zap.Uint32("userId", r.GetUserId()))
	}
	if r, ok := value.(interface{ GetRoundId() string }); ok {
		fields = append(fields, zap.String("roundId", r.GetRoundId()))
	}
	if r, ok := value.(interface{ GetCurrencyType() string }); ok {
		fields = append(fields, zap.String("currencyType", r.GetCurrencyType()))
	}
	rv := reflect.ValueOf(value)
	for rv.IsValid() && rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return fields
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return fields
	}
	allowed := []struct{ name, key string }{{"AgentID", "agentId"}, {"AgentId", "agentId"}, {"GroupID", "groupId"}, {"OperationID", "operationId"}, {"RequestID", "requestId"}, {"BatchID", "batchId"}, {"OperatorID", "operatorId"}, {"Kind", "kind"}, {"Version", "version"}, {"Revision", "revision"}, {"LastOperationID", "lastOperationId"}, {"SourceSymbol", "sourceSymbol"}, {"Symbol", "symbol"}, {"PoolSymbol", "poolSymbol"}, {"RoundID", "roundId"}, {"RoundId", "roundId"}, {"UserID", "userId"}, {"UserId", "userId"}, {"GameID", "gameId"}, {"GameId", "gameId"}}
	for _, item := range allowed {
		field := rv.FieldByName(item.name)
		if !field.IsValid() {
			continue
		}
		switch field.Kind() {
		case reflect.String:
			fields = append(fields, zap.String(item.key, field.String()))
		case reflect.Int, reflect.Int32, reflect.Int64:
			fields = append(fields, zap.Int64(item.key, field.Int()))
		case reflect.Uint, reflect.Uint32, reflect.Uint64:
			fields = append(fields, zap.Uint64(item.key, field.Uint()))
		}
	}
	if id := rv.FieldByName("ID"); id.IsValid() && id.Kind() == reflect.Int64 {
		key := "groupId"
		if rv.Type().Name() == "PoolGroupOperation" {
			key = "operationId"
		}
		fields = append(fields, zap.Int64(key, id.Int()))
	}
	requestID, kind := rv.FieldByName("RequestID"), rv.FieldByName("Kind")
	if requestID.IsValid() && kind.IsValid() && requestID.Kind() == reflect.String && kind.Kind() == reflect.String && kind.String() == "reset" {
		fields = append(fields, zap.String("batchId", strings.TrimPrefix(requestID.String(), "reset:")))
	}
	return fields
}
