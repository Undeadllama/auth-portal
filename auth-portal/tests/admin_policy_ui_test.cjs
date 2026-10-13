const vm = require('node:vm'), fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const root = path.join(process.argv[2], 'static/admin'), kind = process.argv[3];
const flush = () => new Promise(resolve => setImmediate(resolve));
async function environment(failed = false, stale = false) {
  const events = {}, requests = [], children = [], fields = {};
  const save = {disabled: false};
  const element = (name) => fields[name] ||= {value: '', checked: false, addEventListener: (event, fn) => events[name + ':' + event] = fn, replaceChildren() {}, append() {}};
  const form = {querySelector: () => save, addEventListener: (event, fn) => events[event] = fn};
  const list = {replaceChildren: () => children.splice(0), append: (...items) => children.push(...items), querySelectorAll: () => children.filter(c => c.checkbox?.checked).map(c => c.checkbox)};
  const prefix = kind === 'consent' ? 'oidc-consent-' : 'oidc-role-';
  const client = {clientId: 'example-client', name: 'Example', scopes: ['openid'], redirectUris: ['https://app.example/callback']};
  const role = {clientId: client.clientId, exportRoles: ['example-reader'], exportPermissions: true};
  if (kind === 'plex') role.plex = {ownerUuid: 'plex-1', sectionKey: '23', matchRole: 'example-library', defaultRole: 'example-reader'};
  const config = kind === 'consent' ? {clients: [{...client, scopes: stale ? ['openid', 'email'] : client.scopes}]} : {clients: [role]};
  const context = vm.createContext({
    Option: function(label, value) {this.value = value;},
    document: {getElementById: id => id === prefix + (kind === 'consent' ? 'form' : 'config-form') ? form : id === prefix + 'clients' ? list : element(id.slice(prefix.length)), createTextNode: value => value, createElement: tag => tag === 'label' ? {append(checkbox) {this.checkbox = checkbox;}} : {}},
    fetch: async (url, options) => {requests.push({url, options}); return {ok: !failed, json: async () => failed ? {error: 'Load failed'} : options.method === 'PUT' ? {ok: true} : url.endsWith('/oauth/clients') ? {clients: [client]} : url.endsWith('/rbac') ? {roles: [{name: 'example-reader'}, {name: 'example-library'}]} : {version: 7, config}};}
  });
  const modules = new Map();
  async function module(name) {if (modules.has(name)) return modules.get(name); const m = new vm.SourceTextModule(fs.readFileSync(path.join(root, name), 'utf8'), {context}); modules.set(name, m); await m.link(spec => module(path.basename(spec))); return m;}
  await (await module(kind === 'consent' ? 'oidc-consent-config.js' : 'oidc-role-config.js')).evaluate(); await flush();
  return {events, requests, fields, save, children, element};
}
(async () => {
  let e = await environment(); assert.equal(e.save.disabled, false); e.element('reason').value = 'Review example policy';
  await e.events.submit({preventDefault() {}});
  const write = e.requests.find(r => r.options.method === 'PUT'); assert.ok(write);
  const body = JSON.parse(write.options.body); assert.equal(body.version, 7); assert.equal(body.reason, 'Review example policy');
  if (kind === 'consent') {assert.deepEqual(body.clientIds, ['example-client']); e = await environment(false, true); assert.equal(e.children.filter(c => c.checkbox?.checked).length, 0);}
  else {const p = body.config.clients[0]; assert.deepEqual(p.exportRoles, ['example-reader']); assert.equal(p.exportPermissions, true); if (kind === 'plex') assert.deepEqual(p.plex, {ownerUuid: 'plex-1', sectionKey: '23', matchRole: 'example-library', defaultRole: 'example-reader'}); else assert.equal('plex' in p, false);}
  e = await environment(true); assert.equal(e.save.disabled, true); assert.equal(e.element('status').textContent, 'Load failed'); assert.equal(e.requests.filter(r => r.options.method === 'PUT').length, 0);
  console.log('Passed: ' + kind + ' policy UI load/save/version/reason, feature-specific payload and failed-load protection.');
})().catch(e => {console.error(e); process.exitCode = 1;});
