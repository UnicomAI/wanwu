// Node-only UI controller tests. No credentials, browser login or provider calls.
const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const compiler = require('vue-template-compiler/build');
function component(file, api, extra = {}) {
  const source = fs.readFileSync(path.join(__dirname, '../src', file), 'utf8');
  const sfc = compiler.parseComponent(source);
  assert.deepEqual(compiler.compile(sfc.template.content).errors, []);
  const script = sfc.script.content.replace(/^import .*?;$/gm, '').replace('export default', 'module.exports =');
  const context = { module: { exports: {} }, URL, URLSearchParams, Date, Boolean, setTimeout: () => 1, clearTimeout: () => {}, ...api, ...extra };
  vm.runInNewContext(script, context);
  const options = context.module.exports;
  const instance = { ...options.data(), providerIds: [], $set: (obj, key, value) => obj[key] = value, $emit: () => {}, $message: { info() {}, success() {} }, $confirm: async () => {} };
  for (const [key, method] of Object.entries(options.methods || {})) instance[key] = method.bind(instance);
  return { instance, options };
}
test('authorize prepares a visible URL, then saves after callback status', async () => {
  const calls = [];
  const { instance } = component('components/McpOAuthPanel.vue', { mcpOAuthAction: async (action) => { calls.push(action); return { data: action === 'prepare' ? { authorizationUrl: 'https://provider.example/authorize?state=mock' } : { authorized: true } }; } });
  let poll;
  const actualPoll = instance.poll; instance.poll = () => {};
  await instance.prepare({ id: 'sample' });
  assert.equal(instance.active, 'sample'); assert.match(instance.authorizationUrl, /^https:/);
  instance.poll = actualPoll; await instance.poll({ id: 'sample' });
  assert.deepEqual(calls, ['prepare', 'status', 'connect']);
  assert.equal(instance.authorized.sample, true); assert.equal(instance.authorizationUrl, '');
});
test('non-HTTPS authorization URL is rejected and no popup is opened', async () => {
  const { instance } = component('components/McpOAuthPanel.vue', { mcpOAuthAction: async () => ({ data: { authorizationUrl: 'javascript:alert(1)' } }) });
  await instance.prepare({ id: 'sample' });
  assert.equal(instance.authorizationUrl, ''); assert.match(instance.messages.sample, /失败/);
});
test('cancel disconnect preserves the grant without a backend call', async () => {
  let count = 0;
  const { instance } = component('components/McpOAuthPanel.vue', { mcpOAuthAction: async () => { count++; } });
  instance.$confirm = async () => { throw new Error('cancel'); };
  await instance.disconnect({ id: 'sample' }); assert.equal(count, 0);
});
test('callback rejects duplicate state and clears URL history', () => {
  let cleared;
  const { instance, options } = component('views/mcpOAuth/callback.vue', {}, { window: { location: { search: '?code=x&state=a&state=b', pathname: '/aibase/oauth/mcp/callback' }, history: { replaceState: (_, __, value) => cleared = value } }, localStorage: { getItem: () => null } });
  instance.$router = { resolve: () => ({ href: '/aibase/login' }) };
  options.mounted.call(instance);
  assert.equal(instance.code, ''); assert.equal(cleared, '/aibase/oauth/mcp/callback');
});
test('callback waits for same-browser login and submits code through authenticated API', async () => {
  let sent;
  const { instance, options } = component('views/mcpOAuth/callback.vue', { mcpOAuthAction: async (action, data) => { sent = { action, ...data }; } }, { window: { location: { search: '?code=mock-code&state=mock-state', pathname: '/aibase/oauth/mcp/callback' }, history: { replaceState() {} } }, localStorage: { getItem: () => null } });
  instance.$router = { resolve: () => ({ href: '/aibase/login' }) };
  options.mounted.call(instance); assert.equal(sent, undefined);
  await instance.finish(); assert.equal(sent.action, 'complete'); assert.equal(sent.code, 'mock-code'); assert.equal(instance.done, true); assert.equal(instance.code, '');
});
