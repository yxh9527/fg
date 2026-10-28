const safeContext = (context = {}) => {
  const fields = {};
  ['agentId', 'groupId', 'requestId', 'version'].forEach(key => {
    if (typeof context[key] === 'string' || typeof context[key] === 'number') fields[key] = context[key];
  });
  return fields;
};

export function tracePoolGroupCall(name, operation, context = {}, level = 'debug') {
  const started = Date.now();
  const fields = { module: 'pool-group', function: name, ...safeContext(context) };
  const write = (stage, extra = {}) => console[level === 'info' ? 'info' : 'debug']('[pool-group]', { ...fields, stage, ...extra });
  write('entry');
  const success = value => { write('exit', { elapsedMs: Date.now() - started }); return value; };
  const failure = error => {
    console.error('[pool-group]', { ...fields, stage: 'error', elapsedMs: Date.now() - started, status: error?.response?.status, message: error?.response?.msg || error?.message || '操作失败' });
    throw error;
  };
  try {
    const result = operation();
    return result && typeof result.then === 'function' ? result.then(success, failure) : success(result);
  } catch (error) { return failure(error); }
}

export function withPoolGroupLogs(component) {
  const wrap = (name, method) => function (...args) {
    return tracePoolGroupCall(`PoolGroupPanel.${name}`, () => method.apply(this, args), { agentId: this.agentId, groupId: this.form?.groupId, version: this.form?.version }, /save|delete/i.test(name) ? 'info' : 'debug');
  };
  Object.keys(component.methods || {}).forEach(name => { component.methods[name] = wrap(name, component.methods[name]); });
  ['mounted', 'beforeDestroy'].forEach(name => { if (typeof component[name] === 'function') component[name] = wrap(name, component[name]); });
  return component;
}
