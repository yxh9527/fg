const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const compiler = require('vue-template-compiler');

const file = path.join(__dirname, '../src/views/control/PoolSchemePanel.vue');
const component = compiler.parseComponent(fs.readFileSync(file, 'utf8'));

function instance(overrides = {}) {
  const api = {
    getScheme: async () => ({ data: { data: { scheme: { version: 0 }, groups: [], schemes: [] } } }),
    getSchemeGames: async () => ({ data: { data: [] } }),
    getAgents: async () => ({ data: { data: { rows: [], total: 0 } } }),
    getBatches: async () => ({ data: { data: { rows: [], total: 0 } } }),
    previewScheme: async () => ({ data: { data: { rows: [], total: 0 } } }),
    previewAgents: async () => ({ data: { data: [] } }),
    getAgentDetail: async () => ({ data: { data: { groups: [] } } }),
    publishScheme: async () => ({ data: { data: { id: 1 } } }),
    submitAgents: async () => ({ data: { data: { id: 1 } } }),
    ...overrides,
  };
  const sandbox = { module: { exports: {} }, poolSchemes: api, PoolGroupPanel: {},
    window: { crypto: require('node:crypto').webcrypto }, Uint32Array,
    setTimeout: () => 1, clearTimeout: () => {} };
  vm.runInNewContext(component.script.content.replace(/^import .*;$/gm, '').replace('export default', 'module.exports ='), sandbox);
  const options = sandbox.module.exports;
  const model = { sites: [{ id: 1, name: '站点一' }], games: [], ...options.data(),
    $set: (array, index, value) => { array[index] = value; }, $refs: {},
    $emit: () => {}, $message: { warning: () => {}, success: () => {} }, $confirm: async () => {} };
  for (const [key, method] of Object.entries(options.methods)) model[key] = method.bind(model);
  for (const [key, getter] of Object.entries(options.computed)) Object.defineProperty(model, key, { get: () => getter.call(model) });
  return model;
}

test('scheme page and its parent templates compile', () => {
  assert.deepEqual(compiler.compile(component.template.content).errors, []);
  const parent = compiler.parseComponent(fs.readFileSync(path.join(__dirname, '../src/views/control/UserControlPage.vue'), 'utf8'));
  assert.deepEqual(compiler.compile(parent.template.content).errors, []);
});

test('initial scope is explicitly global and requires no selected agent', () => {
  const model = instance();
  assert.equal(model.scope, 'all'); assert.equal(model.scopeWebId, 0); assert.equal(model.scopeReady, true);
  assert.equal(model.scopeLabel, '全部站点');
  model.scope = 'site'; assert.equal(model.scopeReady, false);
  model.scopeSiteId = 1; assert.equal(model.scopeWebId, 1); assert.equal(model.scopeLabel, '站点一');
});

test('draft source follows actual first join and remains stable after removing that member', () => {
  const model = instance(); model.editGroup(); model.groupForm.name = '组';
  model.groupSymbolsChanged(['b']); model.groupSymbolsChanged(['a', 'b']); model.saveGroupDraft();
  assert.deepEqual(Array.from(model.groups[0].symbols), ['b', 'a']); assert.equal(model.groups[0].sourceSymbol, 'b');
  model.editGroup(model.groups[0]); model.groupSymbolsChanged(['a']); model.saveGroupDraft();
  assert.equal(model.groups[0].sourceSymbol, 'b'); assert.equal(model.groups[0].symbols[0], 'a');
});

test('a stale scope response cannot overwrite the current scheme', async () => {
  const waiting = {};
  const model = instance({ getScheme: params => new Promise(resolve => { waiting[params.webId] = resolve; }) });
  const first = model.fetchScheme(); model.scope = 'site'; model.scopeSiteId = 1; const second = model.fetchScheme();
  waiting[1]({ data: { data: { scheme: { version: 2 }, groups: [{ key: 'new' }] } } }); await second;
  waiting[0]({ data: { data: { scheme: { version: 1 }, groups: [{ key: 'old' }] } } }); await first;
  assert.equal(model.version, 2); assert.equal(model.groups[0].key, 'new');
});

test('refresh preserves unpublished draft while still updating progress', async () => {
  let reads = 0;
  const model = instance({ getScheme: async () => { reads++; return { data: { data: { scheme: {}, groups: [] } } }; } });
  model.groups = [{ key: 'draft' }]; await model.fetchGroups(); assert.equal(reads, 0); assert.equal(model.groups[0].key, 'draft');
});

test('default precheck freezes the submitted groups and scope', async () => {
  let submitted;
  const model = instance({ publishScheme: async payload => { submitted = payload; return { data: { data: { id: 8 } } }; } });
  model.groups = [{ key: 'g', name: '组', sourceSymbol: 'a', symbols: ['a'] }];
  await model.openDefaultPreview(); model.groups[0].symbols.push('b'); await model.submitPreview();
  assert.deepEqual(Array.from(submitted.groups[0].symbols), ['a']); assert.equal(submitted.scope, 'all'); assert.equal(submitted.webId, 0);
});

test('agent submission includes only executable agents and their checked generations', async () => {
  let submitted;
  const model = instance({ previewAgents: async () => ({ data: { data: [
    { agentId: 7, generation: 3, error: '' }, { agentId: 8, generation: 5, error: '等待结算' },
  ] } }), submitAgents: async payload => { submitted = payload; return { data: { data: { id: 9 } } }; } });
  model.selectedAgentIds = [7, 8]; model.agentAction = 'restore'; await model.openAgentPreview(); await model.submitPreview();
  assert.deepEqual(Array.from(submitted.agentIds), [7]); assert.equal(submitted.expected['7'], 3); assert.equal(submitted.expected['8'], undefined);
});

test('all failed prechecks never submit an agent operation', async () => {
  let writes = 0;
  const model = instance({ previewAgents: async () => ({ data: { data: [{ agentId: 7, error: '冲突' }] } }), submitAgents: async () => { writes++; } });
  model.selectedAgentIds = [7]; model.agentAction = 'keep'; await model.openAgentPreview(); await model.submitPreview(); assert.equal(writes, 0);
});

test('a network retry reuses its request identity and does not claim application success', async () => {
  const requests = []; let calls = 0;
  const model = instance({ publishScheme: async payload => { requests.push(payload); if (++calls === 1) throw new Error('network'); return { data: { data: { id: 1 } } }; } });
  await model.openDefaultPreview(); await model.submitPreview(); assert.equal(model.previewDialog, true);
  await model.submitPreview(); assert.equal(requests[0].requestId, requests[1].requestId); assert.equal(model.saving, false);
});

test('older representative-agent lookup cannot replace the current group options', async () => {
  const waiting = {};
  const model = instance({ getAgentDetail: params => new Promise(resolve => { waiting[params.agentId] = resolve; }) });
  model.selectedAgentIds = [7]; const first = model.loadAgentGroups(); model.selectedAgentIds = [8]; const second = model.loadAgentGroups();
  waiting[8]({ data: { data: { groups: [{ key: 'eight' }] } } }); await second;
  waiting[7]({ data: { data: { groups: [{ key: 'seven' }] } } }); await first;
  assert.equal(model.agentGroupOptions[0].key, 'eight');
});

test('disabled preview explains the empty selection from the reported screenshot', () => {
  const model = instance(); model.view = 'agents'; model.agentWebId = 1; model.search = '代理0'; model.modeFilter = 'migration'; model.gamesReady = true;
  assert.match(model.agentEmptyNotice, /清除筛选/);
  assert.match(model.agentPreviewDisabledReason, /勾选至少一个正式代理/);
  assert.match(model.agentPreviewDisabledReason, /ID 0/);
  model.selectedAgentIds = [7]; assert.equal(model.agentPreviewDisabledReason, '');
});

test('clear filters queries every formal agent without discarding the create draft', async () => {
  let queried;
  const model = instance({ getAgents: async params => { queried = params; return { data: { data: { rows: [], total: 0 } } }; } });
  model.agentWebId = 1; model.search = '代理0'; model.modeFilter = 'migration'; model.agentPage = 3;
  model.agentForm.name = 'test4'; model.agentForm.symbols = ['b', 'a'];
  await model.resetAgentFilters();
  assert.equal(queried.webId, 0); assert.equal(queried.search, ''); assert.equal(queried.mode, undefined); assert.equal(queried.page, 1);
  assert.equal(model.agentForm.name, 'test4'); assert.deepEqual(Array.from(model.agentForm.symbols), ['b', 'a']);
});

test('opening the test-agent shortcut copies a draft and never submits it', () => {
  let writes = 0; const model = instance({ submitAgents: async () => { writes++; } });
  model.view = 'agents'; model.agentForm.name = 'test4'; model.agentForm.symbols = ['b', 'a']; model.openTestAgent();
  assert.equal(model.view, 'test'); assert.equal(model.testAgentDraft.name, 'test4');
  assert.deepEqual(Array.from(model.testAgentDraft.symbols), ['b', 'a']);
  model.testAgentDraft.symbols.push('c'); assert.equal(model.agentForm.symbols.length, 2); assert.equal(writes, 0);
});

test('catalog failure offers recovery and does not block restore or keep actions', async () => {
  let calls = 0;
  const model = instance({ getSchemeGames: async () => { if (++calls === 1) throw new Error('游戏接口不可用'); return { data: { data: [{ symbol: 'a' }] } }; } });
  model.selectedAgentIds = [7]; await model.fetchCatalog();
  assert.equal(model.gamesReady, false); assert.match(model.catalogError, /游戏接口不可用/); assert.match(model.agentPreviewDisabledReason, /重新获取游戏/);
  model.agentAction = 'restore'; assert.equal(model.agentPreviewDisabledReason, '');
  model.agentAction = 'keep'; assert.equal(model.agentPreviewDisabledReason, '');
  model.agentAction = 'custom'; model.agentKind = 'delete'; assert.equal(model.agentPreviewDisabledReason, '');
  await model.fetchCatalog(); assert.equal(model.gamesReady, true); assert.equal(model.catalogError, ''); assert.equal(model.loadingCatalog, false);
});

test('selecting a formal agent after writing the new group keeps the draft', async () => {
  const model = instance(); model.agentForm.name = 'test4'; model.agentForm.symbols = ['b', 'a']; model.selectedAgentIds = [7];
  await model.loadAgentGroups(); assert.equal(model.agentForm.name, 'test4'); assert.deepEqual(Array.from(model.agentForm.symbols), ['b', 'a']);
});

test('agent-list failures are distinguishable from successful empty results', async () => {
  let calls = 0;
  const model = instance({ getAgents: async () => { if (++calls === 1) throw new Error('代理接口不可用'); return { data: { data: { rows: [], total: 0 } } }; } });
  await model.fetchAgents(); assert.match(model.agentLoadError, /代理接口不可用/); assert.match(model.agentEmptyNotice, /加载失败/);
  await model.fetchAgents(); assert.equal(model.agentLoadError, ''); assert.match(model.agentEmptyNotice, /暂无正式代理/);
});

test('1111 is visible in the complete default preview and included when publishing without selecting a group', async () => {
  let prechecked, published;
  const model = instance({ previewScheme: async payload => { prechecked = payload; return { data: { data: { rows: [], total: 0 } } }; }, publishScheme: async payload => { published = payload; return { data: { data: { id: 12 } } }; } });
  const previous = { key: 'test2', name: 'test2', sourceSymbol: 'a', symbols: ['a'] };
  model.groups = [previous]; model.savedGroups = JSON.stringify(model.groups);
  model.editGroup(); model.groupForm.name = '1111'; model.groupSymbolsChanged(['b', 'c']); model.saveGroupDraft();
  assert.equal(model.draftGroupStatus(model.groups[1]), '新增待发布'); assert.equal(model.selectedAgentIds.length, 0);
  await model.openDefaultPreview();
  assert.equal(model.previewDefaultGroupCount, 2);
  assert.equal(model.previewGroupRows[0].publicationStatus, '保留');
  assert.equal(model.previewGroupRows[1].name, '1111'); assert.equal(model.previewGroupRows[1].publicationStatus, '新增待发布');
  assert.deepEqual(Array.from(model.previewGroupRows[1].symbols), ['b', 'c']);
  await model.submitPreview();
  assert.deepEqual(Array.from(prechecked.groups, group => group.name), ['test2', '1111']);
  assert.deepEqual(Array.from(published.groups, group => group.name), ['test2', '1111']);
  assert.equal(published.groups[1].key, prechecked.groups[1].key);
});

test('default preview distinguishes changed, retained and removed groups even with equal group counts', async () => {
  const model = instance();
  model.savedGroups = JSON.stringify([{ key: 'one', name: 'test2', sourceSymbol: 'a', symbols: ['a'] }, { key: 'two', name: 'old', sourceSymbol: 'b', symbols: ['b'] }]);
  model.groups = [{ key: 'one', name: 'test2 changed', sourceSymbol: 'a', symbols: ['a', 'c'] }, { key: 'three', name: '1111', sourceSymbol: 'b', symbols: ['b'] }];
  await model.openDefaultPreview();
  assert.equal(model.previewGroupRows[0].publicationStatus, '修改待发布'); assert.equal(model.previewGroupRows[1].publicationStatus, '新增待发布');
  assert.equal(model.previewRemovedGroups[0].name, 'old'); assert.equal(model.previewDefaultGroupCount, 2);
  model.groups[1].name = 'a later draft'; model.savedGroups = '[]';
  assert.equal(model.previewGroupRows[1].name, '1111'); assert.equal(model.previewRemovedGroups[0].name, 'old');
});

test('publishing an empty default preview lists removed groups and keeps zero groups in the request', async () => {
  const model = instance(); model.savedGroups = JSON.stringify([{ key: 'one', name: 'old', sourceSymbol: 'a', symbols: ['a'] }]);
  await model.openDefaultPreview(); assert.equal(model.previewDefaultGroupCount, 0); assert.equal(model.previewGroupRows.length, 0); assert.equal(model.previewRemovedGroups[0].name, 'old');
});

test('already published 1111 is labeled as retained in the next preview', async () => {
  const model = instance(); model.groups = [{ key: 'one', name: '1111', sourceSymbol: 'a', symbols: ['a'] }]; model.savedGroups = JSON.stringify(model.groups);
  assert.equal(model.draftGroupStatus(model.groups[0]), '已发布'); await model.openDefaultPreview(); assert.equal(model.previewGroupRows[0].publicationStatus, '保留');
});
