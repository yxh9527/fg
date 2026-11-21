const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const compiler = require('vue-template-compiler');

const directory = path.join(__dirname, '../src/views/settlement');
const component = compiler.parseComponent(fs.readFileSync(path.join(directory, 'PandaRecordDetail.vue'), 'utf8'));
const sandbox = { module: { exports: {} }, config: { ossUrl: '/assets/' } };
vm.runInNewContext(fs.readFileSync(path.join(directory, 'settlementHelpers.js'), 'utf8')
  .replace(/^import .*;$/gm, '').replace(/^export /gm, '') +
  '\nmodule.exports = { asImageList, formatAmount, pickLineInfos, normalizeSettlementRow };', sandbox);
const helpers = sandbox.module.exports;
Object.assign(sandbox, helpers);
vm.runInNewContext(component.script.content.replace(/^import .*;$/gm, '')
  .replace('export default', 'module.exports ='), sandbox);
const options = sandbox.module.exports;

function detail(info) {
  const model = { record: helpers.normalizeSettlementRow({ gameId: 2018, roundID: 'panda-1', detail: { info } }, '/assets/') };
  for (const [key, getter] of Object.entries(options.computed)) Object.defineProperty(model, key, { get: () => getter.call(model) });
  for (const [key, method] of Object.entries(options.methods)) model[key] = method.bind(model);
  return model;
}

test('Panda and shared dialog templates compile', () => {
  for (const template of [component.template.content, compiler.parseComponent(fs.readFileSync(path.join(directory, 'SettlementRecordDialog.vue'), 'utf8')).template.content]) {
    assert.deepEqual(compiler.compile(template).errors, []);
  }
});

test('ordinary and scatter wins share a total and use existing 40x40 assets', () => {
  const model = detail({ grids: [[1, 2, 3]], lines_info: [
    { line_id: '01', ts: 1, bonus: '2.00', times: 20, symbol: [7] },
    { line_id: '02', bonus: 0 },
  ], scatter_info: [{ num: 2, bonus: '2.50', times: 1, img_url: '/global/game/game_FG/2018/40x40/card_10.png' }] });
  assert.equal(model.lines.length, 2);
  assert.equal(model.lineBonus, 4.5);
  assert.equal(model.lines[1].special, true);
  assert.equal(model.lineSymbols(model.lines[1]).length, 2);
  assert.equal(model.cardUrl(model.lineSymbols(model.lines[0])[0]), '/assets/global/game/game_FG/2018/40x40/card_07.png?v=1.23');
  assert.equal(model.linePoints(model.lines[0]), '5,15 15,15 25,15 35,15 45,15');
  assert.equal(model.columns[0].length, 3);
});

test('free spin amounts retain parentheses and paid multipliers', () => {
  const model = detail({ type_id: 2, times: 3, line_bets: '(0.10)', all_bets: '(2.50)', lines_info: [{ line_id: '04', ts: 1, times: 12, bonus: '1.20' }] });
  assert.equal(model.summary.find(item => item.label === '线注').value, '(0.10)');
  assert.equal(model.summary.find(item => item.label === '总投注').value, '(2.50)');
  assert.equal(model.lines[0].times, 12);
  assert.equal(model.linePoints(model.lines[0]), '5,5 15,15 25,25 35,15 45,5');
});

test('reward-only rounds display every coin including zero without requiring a board', () => {
  const model = detail({ type_id: 3, grids: [], sum_bonus: '70.00', bonus_game_info: { bonus_info: [
    { bonus_type: '金币奖励', bonus_num: 0 }, { bonus_type: '金币奖励', bonus_num: '70.00' },
  ], xiao_all_bonus: '70.00' } });
  assert.equal(model.columns.length, 0);
  assert.equal(model.bonusItems.length, 2);
  assert.equal(model.formatAmount(model.bonusItems[0].bonus_num), '0.00');
  assert.equal(model.summary.find(item => item.label === '总奖金').value, '70.00');
});

test('visible reel positions wrap correctly at the last row', () => {
  const model = detail({ grids: [[1, 2, 3]], grid_show: [[3, 4, 5]] });
  assert.match(model.columns[0][0], /card_03/);
  assert.match(model.columns[0][1], /card_01/);
  assert.match(model.columns[0][2], /card_02/);
});
