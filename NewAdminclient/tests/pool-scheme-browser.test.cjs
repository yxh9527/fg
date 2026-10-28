const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { pathToFileURL } = require('node:url');
const { spawnSync } = require('node:child_process');
const compiler = require('vue-template-compiler');
const less = require('less');

const root = path.resolve(__dirname, '..');
const browser = [process.env.CHROME_BIN,
  'C:/Program Files/Google/Chrome/Application/chrome.exe',
  'C:/Program Files (x86)/Microsoft/Edge/Application/msedge.exe',
  '/usr/bin/google-chrome', '/usr/bin/chromium', '/usr/bin/chromium-browser',
].find(candidate => candidate && fs.existsSync(candidate));

test('real table supports agent selection and displays and publishes the complete default draft', { skip: browser ? false : 'Set CHROME_BIN to run the browser regression', timeout: 45000 }, async () => {
  const tempRoot = path.join(root, '.tmp'); fs.mkdirSync(tempRoot, { recursive: true });
  const runDir = fs.mkdtempSync(path.join(tempRoot, 'pool-selection-'));
  const component = compiler.parseComponent(fs.readFileSync(path.join(root, 'src/views/control/PoolSchemePanel.vue'), 'utf8'));
  const source = component.script.content.replace(/^import .*;$/gm, '').replace('export default', 'const options =');
  const style = await less.render(fs.readFileSync(path.join(root, 'src/styles/index.less'), 'utf8'));
  const assetURL = relative => pathToFileURL(path.join(root, relative)).href;
  const html = `<!doctype html><html><head><meta charset="utf-8">
    <link rel="stylesheet" href="${assetURL('node_modules/element-ui/lib/theme-chalk/index.css')}"><style>${style.css}</style>
    </head><body><div id="app"></div>
    <script src="${assetURL('node_modules/vue/dist/vue.js')}"></script>
    <script src="${assetURL('node_modules/element-ui/lib/index.js')}"></script><script>
    Vue.use(ELEMENT); const PoolGroupPanel = { template: '<div></div>' };
    const respond = data => Promise.resolve({ data: { data } });
    const previews = [];
    const defaultPreviews = [], publications = [];
    let storedScheme = { scheme: { version: 1 }, groups: [{ key: 'one', name: '组一', sourceSymbol: 'a', symbols: ['a'] }], schemes: [] };
    const poolSchemes = {
      getSchemeGames: () => respond([{ symbol: 'a', name: 'A' }, { symbol: 'b', name: '怪物命运' }, { symbol: 'c', name: '阿施安' }]),
      getScheme: () => respond(JSON.parse(JSON.stringify(storedScheme))),
      getAgents: params => respond({ rows: [{ agentId: params.page === 2 ? 3 : 2, name: params.page === 2 ? 'test4' : 'test3', mode: 'default', status: 'applied' }], total: 31 }),
      getBatches: () => respond({ rows: [], total: 0 }),
      getAgentDetail: () => respond({ groups: [] }),
      previewAgents: payload => { previews.push(payload); return respond(payload.agentIds.map(agentId => ({ agentId, generation: 1, error: '' }))); },
      previewScheme: payload => { defaultPreviews.push(JSON.parse(JSON.stringify(payload))); return respond({ rows: [{ agentId: 2, name: 'test3', changes: ['当前1组，目标2组'], error: '' }], total: 1 }); },
      publishScheme: payload => { publications.push(JSON.parse(JSON.stringify(payload))); storedScheme = { scheme: { version: 2 }, groups: payload.groups, schemes: [] }; return respond({ id: 12 }); },
    };
    ${source}
    options.template = ${JSON.stringify(component.template.content)};
    const root = new Vue({ render: h => h(options) }).$mount('#app');
    const settle = () => new Promise(resolve => setTimeout(resolve, 80));
    const selection = panel => panel.selectedAgentIds.slice();
    const checkbox = table => table.$el.querySelector('.el-table__body input[type=checkbox]');
    const rowButton = table => Array.from(table.$el.querySelectorAll('.el-table__body button')).find(button => ['选择', '取消选择'].includes(button.textContent.trim()));
    (async () => {
      await settle(); const panel = root.$children[0]; const report = { switches: [] };
      for (let index = 0; index < 3; index++) {
        panel.view = 'agents'; await settle();
        const table = panel.$refs.agentTable; const input = checkbox(table);
        report.switches.push({ type: table.store.states.columns[0].type, visible: !!input && input.closest('label').getBoundingClientRect().width > 0 });
        if (index < 2) { panel.view = 'default'; await settle(); }
      }
      let table = panel.$refs.agentTable;
      panel.agentForm.name = '新增组草稿'; panel.agentForm.symbols = ['a'];
      checkbox(table).click(); await settle(); report.checkboxSelection = selection(panel); report.draftAfterSelect = panel.agentForm.name;
      await panel.fetchAgents(); await settle(); report.afterRefresh = selection(panel); report.refreshChecked = checkbox(table).checked;
      panel.view = 'default'; await settle(); panel.view = 'agents'; await settle();
      table = panel.$refs.agentTable; report.afterTabReturn = selection(panel); report.tabReturnChecked = checkbox(table).checked;
      panel.agentPage = 2; await panel.fetchAgents(); await settle(); rowButton(table).click(); await settle(); report.crossPageSelection = selection(panel);
      panel.agentPage = 1; await panel.fetchAgents(); await settle(); report.returnPageChecked = checkbox(table).checked;
      panel.clearSelection(); await settle(); report.afterClear = selection(panel);
      rowButton(table).click(); await settle(); report.buttonSelection = selection(panel); report.buttonLabel = rowButton(table).textContent.trim();
      const precheck = Array.from(panel.$el.querySelectorAll('button')).find(button => button.textContent.trim() === '预检查指定代理');
      report.precheckDisabled = precheck.disabled; panel.agentForm.name = '新增组草稿'; panel.agentForm.symbols = ['a']; precheck.click(); await settle();
      report.previewAgents = previews.length ? previews[0].agentIds : []; report.previewDialog = panel.previewDialog;
      panel.previewDialog = false; await settle(); rowButton(table).click(); await settle(); report.afterCancel = selection(panel);
      panel.view = 'default'; await settle(); panel.editGroup(); await settle();
      panel.groupForm.name = '1111'; panel.groupForm.symbols = ['b', 'c']; await settle();
      Array.from(panel.$el.querySelectorAll('button')).find(button => button.textContent.trim() === '保存到待发布方案').click(); await settle();
      report.groupDraftNames = panel.groups.map(group => group.name); report.draftStatus = panel.draftGroupStatus(panel.groups[1]);
      Array.from(panel.$el.querySelectorAll('button')).find(button => button.textContent.trim() === '预检查并发布').click(); await settle();
      const previewSection = panel.$el.querySelector('[data-testid="default-publish-groups"]');
      report.previewShows1111 = previewSection.textContent.includes('1111');
      report.previewShowsMembers = previewSection.textContent.includes('怪物命运') && previewSection.textContent.includes('阿施安');
      report.previewShowsCount = previewSection.textContent.includes('2 个水池组');
      const publishButton = Array.from(previewSection.closest('.el-dialog').querySelectorAll('button')).find(button => button.textContent.trim() === '发布完整方案（2组）并同步');
      report.publishButtonEnabled = !!publishButton && !publishButton.disabled;
      report.defaultPrecheckedGroups = defaultPreviews[0].groups.map(group => group.name);
      publishButton.click(); await settle(); report.defaultPublishedGroups = publications[0].groups.map(group => group.name);
      report.publishedStatus = panel.draftGroupStatus(panel.groups[1]); report.published1111Symbols = publications[0].groups[1].symbols;
      panel.$destroy(); const result = document.createElement('pre'); result.id = 'selection-report'; result.textContent = JSON.stringify(report); document.body.appendChild(result);
    })().catch(error => { const result = document.createElement('pre'); result.id = 'selection-report'; result.textContent = JSON.stringify({ error: error.stack }); document.body.appendChild(result); });
    </script></body></html>`;
  const htmlPath = path.join(runDir, 'index.html'); fs.writeFileSync(htmlPath, html);
  const result = spawnSync(browser, ['--headless', '--disable-gpu', '--disable-gpu-sandbox', '--no-sandbox', '--no-first-run', '--no-default-browser-check',
    '--disable-background-networking', '--disable-component-update', '--disable-sync',
    `--user-data-dir=${path.join(runDir, 'profile')}`, '--dump-dom', '--virtual-time-budget=5000', pathToFileURL(htmlPath).href,
  ], { encoding: 'utf8', windowsHide: true, timeout: 35000, maxBuffer: 2 * 1024 * 1024 });
  assert.equal(result.error, undefined, result.error?.message);
  assert.equal(result.status, 0, result.stderr);
  const match = result.stdout.match(/<pre id="selection-report">([^<]*)<\/pre>/);
  assert.ok(match, 'Browser did not produce its selection report');
  const report = JSON.parse(match[1]); assert.equal(report.error, undefined, report.error);
  assert.deepEqual(report.switches, Array.from({ length: 3 }, () => ({ type: 'selection', visible: true })));
  assert.deepEqual(report.checkboxSelection, [2]); assert.equal(report.draftAfterSelect, '新增组草稿');
  assert.deepEqual(report.afterRefresh, [2]); assert.equal(report.refreshChecked, true);
  assert.deepEqual(report.afterTabReturn, [2]); assert.equal(report.tabReturnChecked, true);
  assert.deepEqual(report.crossPageSelection, [2, 3]); assert.equal(report.returnPageChecked, true);
  assert.deepEqual(report.afterClear, []); assert.deepEqual(report.buttonSelection, [2]); assert.equal(report.buttonLabel, '取消选择');
  assert.equal(report.precheckDisabled, false); assert.deepEqual(report.previewAgents, [2]); assert.equal(report.previewDialog, true);
  assert.deepEqual(report.afterCancel, []);
  assert.deepEqual(report.groupDraftNames, ['组一', '1111']); assert.equal(report.draftStatus, '新增待发布');
  assert.equal(report.previewShows1111, true); assert.equal(report.previewShowsMembers, true); assert.equal(report.previewShowsCount, true);
  assert.equal(report.publishButtonEnabled, true); assert.deepEqual(report.defaultPrecheckedGroups, ['组一', '1111']);
  assert.deepEqual(report.defaultPublishedGroups, ['组一', '1111']); assert.deepEqual(report.published1111Symbols, ['b', 'c']); assert.equal(report.publishedStatus, '已发布');
});
