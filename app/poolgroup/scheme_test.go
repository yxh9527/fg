package poolgroup

import (
	"app/config"
	"app/tables/manager"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDBNumber atomic.Int64

func schemeTestStore(t *testing.T) Store {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:pool-scheme-%d?mode=memory&cache=shared", testDBNumber.Add(1))), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err = db.AutoMigrate(&manager.Agent{}, &manager.Web{}, &manager.Game{}); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(db); err != nil {
		t.Fatal(err)
	}
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	previous := config.CfgIns
	config.CfgIns = &config.Configs{Lock: &sync.RWMutex{}, Pool: &config.PoolMgr{Default: map[string]*config.Pool{}, Agent: map[string]*config.Pool{}}}
	t.Cleanup(func() { config.CfgIns = previous })
	for i, symbol := range []string{"a", "b", "c"} {
		if err = db.Create(&manager.Game{Id: int64(i + 1), Name: symbol, ConfName: symbol}).Error; err != nil {
			t.Fatal(err)
		}
		config.CfgIns.SetDefaultPool(symbol, &config.Pool{Symbol: symbol, Pool: map[int32]*config.PoolItem{1: {Base: decimal.NewFromInt(100 + int64(i))}}})
	}
	if err = db.Create(&manager.Web{Id: 1, NickName: "site"}).Error; err != nil {
		t.Fatal(err)
	}
	return Store{DB: db, Redis: rdb}
}

func createSchemeAgent(t *testing.T, s Store, id int64, register bool) {
	t.Helper()
	agent := manager.Agent{Id: id, WebId: 1, NickName: fmt.Sprintf("agent-%d", id)}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&agent).Error; err != nil {
			return err
		}
		if register {
			return s.RegisterAgent(tx, agent)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func publishTestScheme(t *testing.T, s Store, version int64, defs []Definition) *manager.PoolSyncBatch {
	t.Helper()
	batch, err := s.PublishScheme(context.Background(), SchemeRequest{Scope: "all", Version: version, Groups: defs, RequestID: fmt.Sprintf("publish-%d", version)}, 1)
	if err != nil {
		t.Fatal(err)
	}
	return batch
}

func executeSchemeOperation(t *testing.T, s Store, op manager.PoolGroupOperation) {
	t.Helper()
	ctx := context.Background()
	lease, err := Acquire(ctx, s.Redis, op.AgentID)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	state, err := s.Load(ctx, op.AgentID)
	if err != nil {
		t.Fatal(err)
	}
	var plan Plan
	if err = json.Unmarshal([]byte(op.Payload), &plan); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Apply(ctx, lease, state, op, plan); err != nil {
		t.Fatal(err)
	}
	if err = s.Finalize(ctx, op, plan); err != nil {
		t.Fatal(err)
	}
}

func drainSchemes(t *testing.T, s Store) {
	t.Helper()
	var cursor int64
	for round := 0; round < 15; round++ {
		if err := s.schemeWorkerPass(context.Background(), &cursor); err != nil {
			t.Fatal(err)
		}
		var ops []manager.PoolGroupOperation
		if err := s.DB.Where("status <> ?", Applied).Order("id ASC").Find(&ops).Error; err != nil {
			t.Fatal(err)
		}
		for _, op := range ops {
			executeSchemeOperation(t, s, op)
		}
		var active, batches int64
		s.DB.Model(&manager.PoolSyncItem{}).Where("status IN ?", []string{Pending, Waiting, Applying, Failed}).Count(&active)
		s.DB.Model(&manager.PoolSyncBatch{}).Where("dispatch_done = ?", false).Count(&batches)
		if active == 0 && batches == 0 {
			return
		}
	}
	var items []manager.PoolSyncItem
	s.DB.Find(&items)
	t.Fatalf("sync did not converge: %+v", items)
}

func submitTestAgents(t *testing.T, s Store, input AgentSchemeRequest) *manager.PoolSyncBatch {
	t.Helper()
	rows, err := s.PreviewAgents(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	input.Expected = map[string]int64{}
	for _, row := range rows {
		if row.Error != "" {
			t.Fatal(row.Error)
		}
		input.Expected[strconv.FormatInt(row.AgentID, 10)] = row.Generation
	}
	if input.RequestID == "" {
		input.RequestID = fmt.Sprintf("agent-request-%d", testDBNumber.Add(1))
	}
	batch, err := s.SubmitAgents(context.Background(), input, 1)
	if err != nil {
		t.Fatal(err)
	}
	return batch
}

func bindingFor(t *testing.T, s Store, id int64) manager.PoolAgentBinding {
	t.Helper()
	var b manager.PoolAgentBinding
	if err := s.DB.First(&b, "agent_id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return b
}
func actualFor(t *testing.T, s Store, id int64) []*config.PoolGroupSnapshot {
	t.Helper()
	g, err := s.sqlGroups(s.DB, id)
	if err != nil {
		t.Fatal(err)
	}
	return g
}
func definition(key string, symbols ...string) Definition {
	return Definition{Key: key, Name: key, Symbols: symbols, SourceSymbol: symbols[0]}
}

func TestSchemeFollowsExistingAndFutureAgentsAndKeepsAgentBalances(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	createSchemeAgent(t, s, 8, false)
	config.CfgIns.SetAgentPool("8", &config.Pool{Symbol: "a", Pool: map[int32]*config.PoolItem{1: {Base: decimal.NewFromInt(800)}}})
	defs := []Definition{definition("shared", "a", "b")}
	publishTestScheme(t, s, 0, defs)
	drainSchemes(t, s)
	seven, eight := actualFor(t, s, 7)[0], actualFor(t, s, 8)[0]
	if seven.ID == eight.ID || seven.Config.Pool[1].Base.String() != "100" || eight.Config.Pool[1].Base.String() != "800" {
		t.Fatal("agent configuration leaked", seven, eight)
	}
	seed(t, s, context.Background(), 7, config.PoolGroupSymbol(seven.ID), 123)
	submitTestAgents(t, s, AgentSchemeRequest{AgentIDs: []int64{8}, Action: "custom", Kind: "update", Group: definition("shared", "a")})
	drainSchemes(t, s)
	defs[0].Symbols = []string{"a", "b", "c"}
	publishTestScheme(t, s, 1, defs)
	drainSchemes(t, s)
	if len(actualFor(t, s, 7)[0].Symbols) != 3 || len(actualFor(t, s, 8)[0].Symbols) != 1 || bindingFor(t, s, 8).Mode != UseCustom {
		t.Fatal("default overwrote custom configuration")
	}
	exists(t, s, context.Background(), 7, config.PoolGroupSymbol(seven.ID), true)
	createSchemeAgent(t, s, 9, true)
	if err := s.CheckInitialization(context.Background(), 9, "c"); err == nil {
		t.Fatal("new agent accepted an uninitialized game")
	}
	if err := s.CheckInitialization(context.Background(), 9, "unrelated"); err != nil {
		t.Fatal("unrelated game blocked", err)
	}
	drainSchemes(t, s)
	if err := s.CheckInitialization(context.Background(), 9, "c"); err != nil {
		t.Fatal(err)
	}
	if len(actualFor(t, s, 9)[0].Symbols) != 3 || bindingFor(t, s, 9).AppliedRevision != bindingFor(t, s, 7).AppliedRevision {
		t.Fatal("new agent missed latest default")
	}
	submitTestAgents(t, s, AgentSchemeRequest{AgentIDs: []int64{8}, Action: "restore"})
	drainSchemes(t, s)
	if actualFor(t, s, 8)[0].ID != eight.ID || len(actualFor(t, s, 8)[0].Symbols) != 3 || bindingFor(t, s, 8).Mode != FollowDefault {
		t.Fatal("restore recreated an existing group or missed latest version")
	}
}

func TestNewAgentUsesSavedSourceAfterSourceRemovedFromMembers(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	defs := []Definition{definition("shared", "a", "b")}
	publishTestScheme(t, s, 0, defs)
	drainSchemes(t, s)
	defs[0].Symbols = []string{"b"}
	publishTestScheme(t, s, 1, defs)
	createSchemeAgent(t, s, 8, true)
	drainSchemes(t, s)
	g := actualFor(t, s, 8)[0]
	if g.SourceSymbol != "a" || g.Config.Pool[1].Base.String() != "100" {
		t.Fatal("saved source was lost", g)
	}
}

func TestSchemeMigratesOnlyMatchingGroupsAndKeepsLegacyOnExplicitChoice(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	op, err := s.Submit(context.Background(), Intent{AgentID: 7, Kind: "create", RequestID: "legacy", Name: "legacy", Symbols: []string{"a"}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	executeSchemeOperation(t, s, *op)
	old := actualFor(t, s, 7)[0]
	seed(t, s, context.Background(), 7, config.PoolGroupSymbol(old.ID), 77)
	publishTestScheme(t, s, 0, []Definition{definition("default", "a", "b")})
	drainSchemes(t, s)
	if bindingFor(t, s, 7).Mode != NeedsMigration || len(actualFor(t, s, 7)[0].Symbols) != 1 {
		t.Fatal("legacy configuration was overwritten without a decision")
	}
	submitTestAgents(t, s, AgentSchemeRequest{AgentIDs: []int64{7}, Action: "keep"})
	drainSchemes(t, s)
	if bindingFor(t, s, 7).Mode != UseCustom || actualFor(t, s, 7)[0].ID != old.ID {
		t.Fatal("keeping legacy changed runtime group")
	}
	exists(t, s, context.Background(), 7, config.PoolGroupSymbol(old.ID), true)
	publishTestScheme(t, s, 1, []Definition{})
	drainSchemes(t, s)
	if len(actualFor(t, s, 7)) != 1 {
		t.Fatal("empty default removed custom group")
	}
}

func TestMatchingLegacyAdoptionDoesNotCreateRuntimeOperation(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	defs := []Definition{definition("shared", "a", "b")}
	op, err := s.Submit(context.Background(), Intent{AgentID: 7, Kind: "create", RequestID: "legacy", Name: "shared", Symbols: []string{"a", "b"}}, 1)
	if err != nil {
		t.Fatal(err)
	}
	executeSchemeOperation(t, s, *op)
	publishTestScheme(t, s, 0, defs)
	drainSchemes(t, s)
	var count int64
	s.DB.Model(&manager.PoolGroupOperation{}).Count(&count)
	if count != 1 || bindingFor(t, s, 7).AppliedRevision == 0 {
		t.Fatal("adoption recreated runtime state", count)
	}
}

func TestLayoutSwapsMembersAtomicallyAndRetryPreservesNewCounters(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	defs := []Definition{definition("one", "a"), definition("two", "b")}
	publishTestScheme(t, s, 0, defs)
	drainSchemes(t, s)
	old := actualFor(t, s, 7)
	for _, g := range old {
		seed(t, s, context.Background(), 7, config.PoolGroupSymbol(g.ID), 55)
	}
	defs[0].Symbols = []string{"b"}
	defs[1].Symbols = []string{"a"}
	publishTestScheme(t, s, 1, defs)
	drainSchemes(t, s)
	current := actualFor(t, s, 7)
	if current[0].ID != old[0].ID || current[1].ID != old[1].ID || current[0].Symbols[0] != "b" {
		t.Fatal("cross-group move lost identity", current)
	}
	var op manager.PoolGroupOperation
	if err := s.DB.Where("agent_id = ?", 7).Order("id DESC").First(&op).Error; err != nil {
		t.Fatal(err)
	}
	seed(t, s, context.Background(), 7, "a", 999)
	executeSchemeOperation(t, s, op)
	exists(t, s, context.Background(), 7, "a", true)
	for _, g := range old {
		exists(t, s, context.Background(), 7, config.PoolGroupSymbol(g.ID), true)
	}
}

func TestNewerPublishedVersionSupersedesUnstartedTargets(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	first := publishTestScheme(t, s, 0, []Definition{definition("one", "a")})
	if err := s.DispatchSchemeBatch(context.Background(), first.ID); err != nil {
		t.Fatal(err)
	}
	latest := publishTestScheme(t, s, 1, []Definition{definition("two", "b")})
	drainSchemes(t, s)
	binding := bindingFor(t, s, 7)
	if binding.AppliedRevision != latest.RevisionID || actualFor(t, s, 7)[0].Symbols[0] != "b" {
		t.Fatal("old version won", binding)
	}
	var oldItems []manager.PoolSyncItem
	s.DB.Where("revision_id = ?", first.RevisionID).Find(&oldItems)
	for _, item := range oldItems {
		if item.Status != Superseded {
			t.Fatal("old task not superseded", item)
		}
	}
}

func TestSchemePublishIdempotencyAndVersionScopeChecks(t *testing.T) {
	s := schemeTestStore(t)
	input := SchemeRequest{Scope: "site", WebID: 1, Groups: []Definition{definition("one", "a")}, RequestID: "request"}
	first, err := s.PublishScheme(context.Background(), input, 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.PublishScheme(context.Background(), input, 1)
	if err != nil || first.ID != second.ID {
		t.Fatal("retry created another batch", second, err)
	}
	input.Groups[0].Name = "different"
	if _, err = s.PublishScheme(context.Background(), input, 1); err == nil {
		t.Fatal("request ID content conflict accepted")
	}
	if _, err = s.PublishScheme(context.Background(), SchemeRequest{Scope: "all", Groups: []Definition{}, RequestID: "overlap"}, 1); err == nil {
		t.Fatal("overlapping global scope accepted")
	}
	input.RequestID = "stale"
	if _, err = s.PublishScheme(context.Background(), input, 1); err == nil {
		t.Fatal("stale version accepted")
	}
}

func TestCustomSubmitRecordsStalePreviewAsConflictWithoutOverwriting(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	publishTestScheme(t, s, 0, []Definition{definition("one", "a")})
	drainSchemes(t, s)
	input := AgentSchemeRequest{AgentIDs: []int64{7}, Action: "keep", RequestID: "old-preview", Expected: map[string]int64{"7": 0}}
	before := bindingFor(t, s, 7)
	batch, err := s.SubmitAgents(context.Background(), input, 1)
	if err != nil {
		t.Fatal(err)
	}
	var item manager.PoolSyncItem
	s.DB.Where("batch_id = ?", batch.ID).First(&item)
	if item.Status != Conflict || !reflect.DeepEqual(before, bindingFor(t, s, 7)) {
		t.Fatal("stale preview overwrote binding", item)
	}
}

func TestInitializationFailureRollsBackPartialGroupsAndCanRecover(t *testing.T) {
	s := schemeTestStore(t)
	publishTestScheme(t, s, 0, []Definition{definition("one", "a"), definition("two", "b")})
	createSchemeAgent(t, s, 7, true)
	// Force task insertion to fail after actual group rows were allocated.
	if err := s.DB.Exec("CREATE TRIGGER fail_layout_insert BEFORE INSERT ON pool_group_operation BEGIN SELECT RAISE(FAIL, 'temporary insertion failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	var cursor int64
	if err := s.schemeWorkerPass(context.Background(), &cursor); err != nil {
		t.Fatal(err)
	}
	var count int64
	s.DB.Model(&manager.PoolGroup{}).Count(&count)
	if count != 0 {
		t.Fatal("failed planning left orphan groups", count)
	}
	if err := s.CheckInitialization(context.Background(), 7, "a"); err == nil {
		t.Fatal("failed initialization reopened games")
	}
	if err := s.DB.Exec("DROP TRIGGER fail_layout_insert").Error; err != nil {
		t.Fatal(err)
	}
	drainSchemes(t, s)
	if len(actualFor(t, s, 7)) != 2 || bindingFor(t, s, 7).Initializing {
		t.Fatal("failed initialization did not recover")
	}
}

func TestAgentAndInitializationTaskRollbackTogether(t *testing.T) {
	s := schemeTestStore(t)
	publishTestScheme(t, s, 0, []Definition{definition("one", "a")})
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		agent := manager.Agent{Id: 7, WebId: 1}
		if err := tx.Create(&agent).Error; err != nil {
			return err
		}
		if err := s.RegisterAgent(tx, agent); err != nil {
			return err
		}
		return fmt.Errorf("rollback")
	})
	if err == nil {
		t.Fatal("expected rollback")
	}
	for _, model := range []interface{}{&manager.Agent{}, &manager.PoolAgentBinding{}, &manager.PoolSyncItem{}} {
		var n int64
		if err := s.DB.Model(model).Count(&n).Error; err != nil || n != 0 {
			t.Fatal("partial creation survived transaction", model, n, err)
		}
	}
}

func TestDefinitionsRejectDuplicateMembershipAndImmutableSourceChanges(t *testing.T) {
	if err := ValidateDefinitions([]Definition{definition("one", "a"), definition("two", "a")}); err == nil {
		t.Fatal("duplicate membership accepted")
	}
	if err := ValidateDefinitions([]Definition{}); err != nil {
		t.Fatal("empty layout must be valid", err)
	}
	previous := []Definition{definition("one", "a", "b")}
	changed := []Definition{definition("one", "b")}
	if err := validateDefinitionSources(changed, previous); err == nil {
		t.Fatal("source changed silently")
	}
}

func TestLayoutGateExcludesUnchangedGroups(t *testing.T) {
	a, b := testGroup(1, 7, "a"), testGroup(2, 7, "b")
	updated := *a
	updated.Symbols = []string{"a", "c"}
	updated.Version++
	gate := OperationGate(&State{Groups: []*config.PoolGroupSnapshot{a, b}}, manager.PoolGroupOperation{ID: 3, AgentID: 7, Kind: "layout"}, Plan{Layout: []*config.PoolGroupSnapshot{&updated, b}})
	if !gate.Affects("a") || !gate.Affects("c") || gate.Affects("b") {
		t.Fatal("layout paused an unrelated group", gate)
	}
}

func TestCompensationRecreatesMissingSyncItemAndStillConverges(t *testing.T) {
	s := schemeTestStore(t)
	publishTestScheme(t, s, 0, []Definition{definition("one", "a")})
	createSchemeAgent(t, s, 7, true)
	if err := s.DB.Where("agent_id = ?", 7).Delete(&manager.PoolSyncItem{}).Error; err != nil {
		t.Fatal(err)
	}
	drainSchemes(t, s)
	if bindingFor(t, s, 7).AppliedRevision == 0 || len(actualFor(t, s, 7)) != 1 {
		t.Fatal("missing initialization task was not compensated")
	}
}

func TestRunningOldLayoutCompletesWithoutAcknowledgingNewTarget(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	first := publishTestScheme(t, s, 0, []Definition{definition("one", "a")})
	var cursor int64
	if err := s.schemeWorkerPass(context.Background(), &cursor); err != nil {
		t.Fatal(err)
	}
	var old manager.PoolGroupOperation
	if err := s.DB.Where("agent_id = ?", 7).First(&old).Error; err != nil {
		t.Fatal(err)
	}
	latest := publishTestScheme(t, s, 1, []Definition{definition("one", "a", "b")})
	if err := s.schemeWorkerPass(context.Background(), &cursor); err != nil {
		t.Fatal(err)
	}
	executeSchemeOperation(t, s, old)
	var oldItems []manager.PoolSyncItem
	s.DB.Where("revision_id = ?", first.RevisionID).Find(&oldItems)
	for _, item := range oldItems {
		if err := s.ProcessSyncItem(context.Background(), item.ID); err != nil {
			t.Fatal(err)
		}
	}
	b := bindingFor(t, s, 7)
	if b.TargetRevision != latest.RevisionID || b.AppliedRevision == latest.RevisionID {
		t.Fatal("old operation acknowledged the newer target", b)
	}
	oldID := actualFor(t, s, 7)[0].ID
	seed(t, s, context.Background(), 7, config.PoolGroupSymbol(oldID), 99)
	drainSchemes(t, s)
	if actualFor(t, s, 7)[0].ID != oldID || len(actualFor(t, s, 7)[0].Symbols) != 2 {
		t.Fatal("latest target was not reached while retaining identity")
	}
	exists(t, s, context.Background(), 7, config.PoolGroupSymbol(oldID), true)
}

func TestEmptyCustomLayoutStaysEmptyWhenDefaultAddsGroups(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	publishTestScheme(t, s, 0, []Definition{definition("one", "a")})
	drainSchemes(t, s)
	submitTestAgents(t, s, AgentSchemeRequest{AgentIDs: []int64{7}, Action: "custom", Kind: "delete", Group: Definition{Key: "one"}})
	drainSchemes(t, s)
	publishTestScheme(t, s, 1, []Definition{definition("one", "a"), definition("two", "b")})
	createSchemeAgent(t, s, 8, true)
	drainSchemes(t, s)
	if len(actualFor(t, s, 7)) != 0 || bindingFor(t, s, 7).Mode != UseCustom || len(actualFor(t, s, 8)) != 2 {
		t.Fatal("empty custom layout inherited new default groups")
	}
}

func TestPublishedPaginationAndLaterRegistrationDoNotLoseAgents(t *testing.T) {
	s := schemeTestStore(t)
	for id := int64(1); id <= 105; id++ {
		createSchemeAgent(t, s, id, false)
	}
	batch := publishTestScheme(t, s, 0, []Definition{definition("one", "a")})
	if err := s.DispatchSchemeBatch(context.Background(), batch.ID); err != nil {
		t.Fatal(err)
	}
	var saved manager.PoolSyncBatch
	s.DB.First(&saved, batch.ID)
	if saved.Cursor != 100 || saved.DispatchDone {
		t.Fatal("dispatch did not retain its page cursor", saved)
	}
	createSchemeAgent(t, s, 106, true)
	if err := s.DispatchSchemeBatch(context.Background(), batch.ID); err != nil {
		t.Fatal(err)
	}
	var n int64
	s.DB.Model(&manager.PoolSyncItem{}).Where("batch_id = ?", batch.ID).Count(&n)
	if n != 105 {
		t.Fatal("published snapshot expanded or lost agents", n)
	}
	if bindingFor(t, s, 106).TargetRevision != batch.RevisionID {
		t.Fatal("later agent missed registration target")
	}
	if err := s.CheckInitialization(context.Background(), 106, "a"); err == nil {
		t.Fatal("later agent lacked initialization gate")
	}
}

func TestMissingRuntimeCannotBeAcknowledgedAsUnchangedSuccess(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 7, false)
	publishTestScheme(t, s, 0, []Definition{definition("one", "a")})
	drainSchemes(t, s)
	if err := s.Redis.Del(context.Background(), Key(7, "state")).Err(); err != nil {
		t.Fatal(err)
	}
	batch := submitTestAgents(t, s, AgentSchemeRequest{AgentIDs: []int64{7}, Action: "keep"})
	var item manager.PoolSyncItem
	if err := s.DB.Where("batch_id = ?", batch.ID).First(&item).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.ProcessSyncItem(context.Background(), item.ID); err != nil {
		t.Fatal(err)
	}
	binding := bindingFor(t, s, 7)
	if binding.Status != Failed || binding.AppliedRevision == binding.TargetRevision {
		t.Fatal("missing runtime was reported applied", binding)
	}
}

func TestPreviewAndPublicationKeep1111WhenGroupCountIsUnchanged(t *testing.T) {
	s := schemeTestStore(t)
	createSchemeAgent(t, s, 2, false)
	before := []Definition{definition("one", "a"), definition("two", "b")}
	before[0].Name = "test2"
	before[1].Name = "old"
	publishTestScheme(t, s, 0, before)
	drainSchemes(t, s)
	actualBefore := actualFor(t, s, 2)
	after := []Definition{before[0], before[1]}
	after[1].Name = "1111"
	after[1].Symbols = []string{"b", "c"}
	preview, err := s.PreviewScheme(context.Background(), SchemeRequest{Scope: "all", Version: 1, Groups: after}, 1, 30)
	if err != nil {
		t.Fatal(err)
	}
	rows := preview["rows"].([]AgentPreview)
	if len(rows) != 1 {
		t.Fatal("missing agent preview", preview)
	}
	description := strings.Join(rows[0].Changes, "；")
	if !strings.Contains(description, "当前2组，目标2组") || !strings.Contains(description, "当前组：test2、old") || !strings.Contains(description, "发布目标组：test2、1111") {
		t.Fatal("equal group counts hid named changes", description)
	}
	if rows[0].Groups[1].Name != "1111" || !reflect.DeepEqual(rows[0].Groups[1].Symbols, []string{"b", "c"}) {
		t.Fatal("preview lost target group", rows[0])
	}
	if actualFor(t, s, 2)[1].Name != "old" {
		t.Fatal("precheck changed applied configuration")
	}
	publishTestScheme(t, s, 1, after)
	drainSchemes(t, s)
	actual := actualFor(t, s, 2)
	if len(actual) != 2 || actual[1].Name != "1111" || actual[1].ID != actualBefore[1].ID || !reflect.DeepEqual(actual[1].Symbols, []string{"b", "c"}) {
		t.Fatal("published 1111 was not applied", actual)
	}
}
