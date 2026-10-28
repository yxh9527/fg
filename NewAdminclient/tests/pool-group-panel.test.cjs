const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const compiler = require('vue-template-compiler');
const source = fs.readFileSync(require('node:path').join(__dirname, '../src/views/control/PoolGroupPanel.vue'), 'utf8');
const component = compiler.parseComponent(source);
function loadLogging(events = []) {
  const loggingSource = fs.readFileSync(require('node:path').join(__dirname, '../src/libs/poolGroupLog.js'), 'utf8');
  const sandbox = {module:{exports:{}},Date,console:Object.fromEntries(['debug','info','error'].map(level=>[level,(...args)=>events.push({level,args})]))};
  vm.runInNewContext(loggingSource.replace(/^export /gm, '') + '\nmodule.exports = { tracePoolGroupCall, withPoolGroupLogs };', sandbox);
  return sandbox.module.exports;
}

function instance(overrides = {}) {
  const emitted = [];
  const api = { getPoolGroups: async () => ({data:{data:{groups:[],operations:[],revision:0}}}), getPoolGroupGames: async () => ({data:{data:[]}}), createPoolGroup: async () => ({}), updatePoolGroup: async () => ({}), deletePoolGroup: async () => ({}) , ...overrides };
  const sandbox = {module:{exports:{}},dayjs:require('dayjs'),window:{crypto:require('node:crypto').webcrypto},Uint32Array,setTimeout:()=>1,clearTimeout:()=>{},...api,...loadLogging()};
  const script = component.script.content.replace(/^import .*;$/gm, '').replace('export default', 'module.exports =');
  vm.runInNewContext(script, sandbox, {filename:'PoolGroupPanel.vue'});
  const options = sandbox.module.exports;
  const model = {agentId:7,siteId:1,games:[],...options.data(),$emit:(...args)=>emitted.push(args),$message:{error:()=>{},success:()=>{}},$confirm:async()=>{}};
  for (const [name,fn] of Object.entries(options.methods)) model[name]=fn.bind(model);
  for (const [name,fn] of Object.entries(options.computed)) Object.defineProperty(model,name,{get:()=>fn.call(model)});
  return {model,emitted};
}
test('component template compiles', () => { assert.deepEqual(compiler.compile(component.template.content).errors, []); });
test('selection keeps actual join order, not catalog order', () => {
  const {model} = instance();model.selectionOrder=['b','a'];model.preserveOrder(['a','c','b']);assert.deepEqual(Array.from(model.form.symbols),['b','a','c']);
  model.preserveOrder(['a','c']);assert.equal(model.sourceSymbol,'a');
});
test('editing members keeps original source and Base even when first member is removed', () => {
  const {model}=instance();model.form={groupId:4,sourceSymbol:'a',symbols:['b'],config:{pool:{1:{base:'100'}}}};model.availableGames=[{symbol:'b',poolConfig:{pool:{1:{base:'999'}}}}];
  assert.equal(model.sourceSymbol,'a');assert.equal(model.previewConfig.base,'100');
});
test('same payload retry reuses request ID; a changed payload gets a new ID', () => {
  const {model}=instance();const payload={agentId:7,name:'G',symbols:['a']};const first=model.nextRequestId(payload);assert.equal(model.nextRequestId(payload),first);assert.notEqual(model.nextRequestId({...payload,symbols:['a','b']}),first);
});
test('new group submits selected order and waits for execution', async () => {
  let submitted;const {model,emitted}=instance({createPoolGroup:async value=>{submitted=value;}});model.form.name='G';model.form.symbols=['b','a'];model.availableGames=[{symbol:'a',poolConfig:{pool:{1:{base:'10'}}}},{symbol:'b',poolConfig:{pool:{1:{base:'20'}}}}];
  await model.saveGroup();assert.deepEqual(Array.from(submitted.symbols),['b','a']);assert.equal(submitted.agentId,7);assert.match(submitted.requestId,/^[a-f0-9]{32}$/);assert.equal(model.saving,false);assert.ok(emitted.some(e=>e[0]==='changed'));
});
test('invalid empty or over-limit membership never reaches API', async () => {
  let calls=0;const {model}=instance({createPoolGroup:async()=>{calls++;}});model.form.name='G';await model.saveGroup();model.form.symbols=Array.from({length:11},(_,i)=>String(i));await model.saveGroup();assert.equal(calls,0);
});
test('stale agent response cannot overwrite current agent results', async () => {
  let finish;const {model}=instance({getPoolGroups:()=>new Promise(resolve=>{finish=resolve;})});const waiting=model.fetchGroups();model.agentId=8;model.sequence++;model.groups=[{id:88}];finish({data:{data:{groups:[{id:77}],operations:[],revision:1}}});await waiting;assert.equal(model.groups[0].id,88);
});
test('method logging records correlation and duration without tokens or raw payload', async () => {
  const events=[];const {tracePoolGroupCall}=loadLogging(events);const result=await tracePoolGroupCall('api.createPoolGroup',async()=>123,{agentId:7,requestId:'r1',token:'secret-token',payload:'raw-private-body'},'info');
  assert.equal(result,123);assert.equal(events[0].args[1].stage,'entry');assert.equal(events[1].args[1].stage,'exit');assert.equal(events[1].args[1].agentId,7);assert.equal(typeof events[1].args[1].elapsedMs,'number');assert.ok(!JSON.stringify(events).includes('secret-token'));assert.ok(!JSON.stringify(events).includes('raw-private-body'));
});
test('failure logging preserves the original error and omits axios request config', async () => {
  const events=[];const {tracePoolGroupCall}=loadLogging(events);const failure={message:'network failed',response:{status:503},config:{params:{token:'secret-token'}}};
  await assert.rejects(tracePoolGroupCall('api.updatePoolGroup',async()=>{throw failure;},{agentId:7}),error=>error===failure);assert.equal(events.at(-1).level,'error');assert.equal(events.at(-1).args[1].status,503);assert.ok(!JSON.stringify(events).includes('secret-token'));
});
