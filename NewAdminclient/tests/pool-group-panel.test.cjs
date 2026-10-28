const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const compiler = require('vue-template-compiler');
const source = fs.readFileSync(require('node:path').join(__dirname, '../src/views/control/PoolGroupPanel.vue'), 'utf8');
const component = compiler.parseComponent(source);
function loadControlHelpers() {
  const helperSource = fs.readFileSync(require('node:path').join(__dirname, '../src/views/control/controlHelpers.js'), 'utf8');
  const sandbox = {module:{exports:{}},dayjs:require('dayjs')};
  vm.runInNewContext(helperSource.replace(/^import .*;$/gm,'').replace(/^export /gm,'') + ';module.exports = {parseOptionalAgentId,formatDateTime,toAmount};', sandbox);
  return sandbox.module.exports;
}
function loadLogging(events = []) {
  const loggingSource = fs.readFileSync(require('node:path').join(__dirname, '../src/libs/poolGroupLog.js'), 'utf8');
  const sandbox = {module:{exports:{}},Date,console:Object.fromEntries(['debug','info','error'].map(level=>[level,(...args)=>events.push({level,args})]))};
  vm.runInNewContext(loggingSource.replace(/^export /gm, '') + '\nmodule.exports = { tracePoolGroupCall, withPoolGroupLogs };', sandbox);
  return sandbox.module.exports;
}

function instance(overrides = {}) {
  const emitted = [];
  const api = { getPoolGroups: async () => ({data:{data:{groups:[],operations:[],revision:0}}}), getPoolGroupGames: async () => ({data:{data:[]}}), createPoolGroup: async () => ({}), updatePoolGroup: async () => ({}), deletePoolGroup: async () => ({}) , ...overrides };
  const sandbox = {module:{exports:{}},dayjs:require('dayjs'),window:{crypto:require('node:crypto').webcrypto},Uint32Array,setTimeout:()=>1,clearTimeout:()=>{},...api,...loadLogging(),...loadControlHelpers()};
  const script = component.script.content.replace(/^import .*;$/gm, '').replace('export default', 'module.exports =');
  vm.runInNewContext(script, sandbox, {filename:'PoolGroupPanel.vue'});
  const options = sandbox.module.exports;
  const model = {agentId:7,siteId:1,games:[],...options.data(),$emit:(...args)=>emitted.push(args),$message:{error:()=>{},warning:()=>{},success:()=>{}},$confirm:async()=>{}};
  for (const [name,fn] of Object.entries(options.methods)) model[name]=fn.bind(model);
  for (const [name,fn] of Object.entries(options.computed)) Object.defineProperty(model,name,{get:()=>fn.call(model)});
  return {model,emitted};
}
function controlPage(query = {}, savedAgent = '') {
  const text = fs.readFileSync(require('node:path').join(__dirname, '../src/views/control/UserControlPage.vue'), 'utf8');
  let importing=false;const body=compiler.parseComponent(text).script.content.split(String.fromCharCode(10)).filter(line=>{if(line.trim().startsWith('import ')) importing=true;if(importing){if(line.includes('from ') && line.trim().endsWith(';')) importing=false;return false;}return true;}).join(String.fromCharCode(10));
  const values=new Map([['siteOption',JSON.stringify([{id:1,name:'站点',agentList:[{id:7,name:'正式代理'}]}])],['games',JSON.stringify([{number:1,name:'A',symbol:'a'}])],['agentVal',savedAgent]]);
  const storage={getItem:key=>values.has(key)?values.get(key):null,setItem:(key,value)=>values.set(key,String(value))};
  const sandbox={module:{exports:{}},dayjs:require('dayjs'),AppTable:{},PoolGroupPanel:{},echarts:{},sessionStorage:storage,...loadControlHelpers()};
  vm.runInNewContext(body.replace('export default','module.exports ='),sandbox,{filename:'UserControlPage.vue'});const options=sandbox.module.exports;
  const model={...options.data(),$route:{query}};for(const [name,method] of Object.entries(options.methods)) model[name]=method.bind(model);for(const [name,getter] of Object.entries(options.computed)) Object.defineProperty(model,name,{get:()=>getter.call(model)});model.refreshActiveTab=()=>{};return {model,storage};
}
test('component template compiles', () => { assert.deepEqual(compiler.compile(component.template.content).errors, []); });
test('a selected agent can open create even when a reset is waiting', () => {
  const {model}=instance();model.operations=[{id:12,kind:'reset',status:'waiting'}];
  const root=compiler.compile(component.template.content).ast;let button;
  const find=node=>{if(node.tag==='el-button' && (node.children || []).some(child=>child.text && child.text.includes('新增水池组'))) button=node;for(const child of node.children || []) find(child);};find(root);
  assert.ok(button);const disabled=vm.runInNewContext(button.attrsMap[':disabled'],{agentId:model.agentId,hasPending:model.hasPending,hasSelectedAgent:model.hasSelectedAgent,saving:model.saving});
  assert.equal(disabled,false,'pending records must not silently disable the create-dialog button');
});
test('create dialog appears immediately while available games are loading', async () => {
  let finish;const {model}=instance({getPoolGroupGames:()=>new Promise(resolve=>{finish=resolve;})});
  const pending=model.openDialog();assert.equal(model.dialogVisible,true);assert.equal(model.loadingGames,true);
  finish({data:{data:[{symbol:'a',poolConfig:{pool:{1:{base:'10'}}}}]}});await pending;
  assert.equal(model.loadingGames,false);assert.equal(model.availableGames[0].symbol,'a');
});
test('pending-task summary explains a blocking operation outside recent history', async () => {
  const {model}=instance({getPoolGroups:async()=>({data:{data:{groups:[],operations:[{id:40,status:'applied'}],pendingOperation:{id:12,kind:'reset',status:'waiting',error:'等待未结算订单'},revision:1}}})});
  await model.fetchGroups();assert.equal(model.hasPending,true);assert.equal(model.pendingTask.id,12);assert.match(model.pendingNotice,/#12/);assert.match(model.pendingNotice,/重置/);assert.match(model.pendingNotice,/等待未结算订单/);
});
test('opening a draft does not bypass the pending-operation submit guard', async () => {
  let calls=0;const {model}=instance({createPoolGroup:async()=>{calls++;}});model.pendingOperation={id:12,kind:'reset',status:'waiting'};model.form.name='G';model.form.symbols=['a'];model.availableGames=[{symbol:'a',poolConfig:{pool:{1:{base:'10'}}}}];
  await model.saveGroup();assert.equal(calls,0);model.pendingOperation.status='applied';await model.saveGroup();assert.equal(calls,1);
});
test('available-game failure stays visible inside the open dialog and can be retried', async () => {
  let calls=0;const {model}=instance({getPoolGroupGames:async()=>{calls++;if(calls===1) throw {response:{msg:'代理不属于当前站点'}};return {data:{data:[{symbol:'a',poolConfig:{pool:{1:{base:'10'}}}}]}};}});
  await model.openDialog();assert.equal(model.dialogVisible,true);assert.equal(model.loadingGames,false);assert.match(model.gameLoadError,/代理不属于当前站点/);model.form.name='保留的草稿';
  await model.fetchDialogGames();assert.equal(model.gameLoadError,'');assert.equal(model.availableGames[0].symbol,'a');assert.equal(model.form.name,'保留的草稿');
});
test('available-game requests from an old agent cannot populate the new dialog', async () => {
  const waiting={};const {model}=instance({getPoolGroupGames:params=>new Promise(resolve=>{waiting[params.agentId]=resolve;})});
  const first=model.openDialog();model.agentId=8;model.resetScope();const second=model.openDialog();
  waiting[8]({data:{data:[{symbol:'new-agent'}]}});await second;waiting[7]({data:{data:[{symbol:'old-agent'}]}});await first;
  assert.equal(model.availableGames[0].symbol,'new-agent');assert.equal(model.loadingGames,false);assert.equal(model.dialogVisible,true);
});
test('zero is a valid test agent while absent or invalid selections stay disabled', () => {
  const {model}=instance();for(const value of [null,undefined,'',-1,'invalid',false,true,' ']){model.agentId=value;assert.equal(model.hasSelectedAgent,false,String(value));}
  for(const value of [0,'0',7,'7']){model.agentId=value;assert.equal(model.hasSelectedAgent,true,String(value));}
});
test('test agent zero can load groups, open a draft and submit its numeric identity', async () => {
  let queried,selected,submitted;const {model}=instance({getPoolGroups:async params=>{queried=params;return {data:{data:{groups:[],operations:[],revision:0}}};},getPoolGroupGames:async params=>{selected=params;return {data:{data:[{symbol:'a',poolConfig:{pool:{1:{base:'10'}}}}]}};},createPoolGroup:async data=>{submitted=data;}});
  model.agentId=0;await model.fetchGroups();assert.equal(queried.agentId,0);await model.openDialog();assert.equal(model.dialogVisible,true);assert.equal(selected.agentId,0);model.form.name='测试组合池';model.form.symbols=['a'];await model.saveGroup();assert.equal(submitted.agentId,0);assert.equal(typeof submitted.agentId,'number');
});

test('test-agent shortcut opens the transferred draft without submitting it', async () => {
  let writes = 0;
  const {model} = instance({ getPoolGroupGames: async () => ({data:{data:[{symbol:'b',poolConfig:{pool:{1:{base:'10'}}}},{symbol:'a',poolConfig:{pool:{1:{base:'20'}}}}]}}), createPoolGroup: async () => { writes++; } });
  model.agentId = 0; model.initialDraft = {name:'test4',symbols:['b','a']}; await model.openInitialDraft();
  assert.equal(model.dialogVisible, true); assert.equal(model.form.name, 'test4'); assert.deepEqual(Array.from(model.form.symbols), ['b','a']);
  assert.equal(model.sourceSymbol, 'b'); assert.equal(writes, 0);
  model.agentId = 7; model.dialogVisible = false; await model.openInitialDraft(); assert.equal(model.dialogVisible, false);
});
test('route agent zero takes precedence over a saved positive agent after refresh', async () => {
  const {model,storage}=controlPage({siteId:'1',agentId:'0'},'7');await model.initBaseData();assert.equal(model.selectedAgentId,0);assert.equal(storage.getItem('agentVal'),'0');assert.equal(model.agentDisplayLabel,'测试代理');assert.ok(model.agentOptions.some(agent=>agent.id===0));
});
test('saved agent zero is restored and clearing selection still means default config', async () => {
  const {model,storage}=controlPage({siteId:'1'},'0');await model.initBaseData();assert.equal(model.selectedAgentId,0);model.handleAgentChange('');assert.equal(model.selectedAgentId,null);assert.equal(storage.getItem('agentVal'),'');assert.equal(model.agentDisplayLabel,'默认配置');model.handleAgentChange('0');assert.equal(model.selectedAgentId,0);assert.equal(storage.getItem('agentVal'),'0');
});
test('cached string zero options normalize to the numeric select value without duplication', async () => {
  const {model,storage}=controlPage({siteId:'1'},'0');storage.setItem('siteOption',JSON.stringify([{id:1,name:'站点',agentList:[{id:'0',name:'原测试代理'},{id:'7',name:'正式代理'}]}]));await model.initBaseData();assert.equal(model.selectedAgentId,0);assert.equal(model.agentOptions.filter(agent=>agent.id===0).length,1);assert.equal(model.agentDisplayLabel,'原测试代理');assert.equal(model.agentOptions[1].id,7);
});
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
