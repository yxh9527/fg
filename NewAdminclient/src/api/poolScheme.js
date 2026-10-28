import { request } from '@/libs/request';
import { getToken } from '@/libs/util';
import { tracePoolGroupCall } from '@/libs/poolGroupLog';

const call = (path, method, values = {}, params = {}) => tracePoolGroupCall(`api.poolScheme.${path || 'get'}`, () => request({
  url: `v2/govern/pool-schemes${path ? '/' + path : ''}`, method,
  params: { token: getToken(), ...(method === 'get' ? values : params) },
  ...(method === 'get' ? {} : { data: values }),
}), values, method === 'get' ? 'debug' : 'info');

export const getScheme = params => call('', 'get', params);
export const getSchemeGames = () => call('games', 'get');
export const previewScheme = (data, params) => call('preview', 'post', data, params);
export const publishScheme = data => call('publish', 'post', data);
export const getAgents = params => call('agents', 'get', params);
export const getAgentDetail = params => call('agent', 'get', params);
export const previewAgents = data => call('agents/preview', 'post', data);
export const submitAgents = data => call('agents/submit', 'post', data);
export const getBatches = params => call('batches', 'get', params);
export const getBatch = params => call('batch', 'get', params);
export const retryBatch = data => call('retry', 'post', data);
