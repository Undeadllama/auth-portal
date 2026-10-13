// node --experimental-vm-modules tests/oidc_config_ui_test.cjs /path/to/source
const vm = require('node:vm');
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const root = path.join(process.argv[2], 'static/admin');

async function mediaEnvironment(failed = false) {
  const events = {}, requests = [], children = [];
  const save = {disabled: false};
  const form = {querySelector: () => save, addEventListener: (name, fn) => events[name] = fn};
  const list = {
    replaceChildren: () => children.splice(0), append: (...items) => children.push(...items),
    querySelectorAll: () => children.filter(c => c.checkbox?.checked).map(c => c.checkbox)
  };
  const elements = {
    'oidc-media-form': form, 'oidc-media-clients': list, 'oidc-media-aliases': {},
    'oidc-media-status': {}, 'oidc-media-reason': {value: 'Test policy'},
    'oidc-media-reload': {addEventListener: (name, fn) => events.reload = fn}
  };
  const context = vm.createContext({
    document: {
      getElementById: id => elements[id], createTextNode: value => value,
      createElement: tag => tag === 'label' ? {append(checkbox) {this.checkbox = checkbox;}} : {}
    },
    fetch: async (url, options) => {
      requests.push({url, options});
      return {ok: !failed, json: async () => failed ? {error: 'Load failed'} : options.method === 'PUT' ? {ok: true} :
        url.endsWith('/clients') ? {clients: [{clientId: 'arbitrary-app', name: 'Example application'}]} :
        {version: 7, config: {clientIds: ['arbitrary-app'], scopeAliases: ['community.member']}}};
    }
  });
  const modules = new Map();
  async function module(name) {
    if (modules.has(name)) return modules.get(name);
    const value = new vm.SourceTextModule(fs.readFileSync(path.join(root, name), 'utf8'), {context});
    modules.set(name, value);
    await value.link(specifier => module(path.basename(specifier)));
    return value;
  }
  await (await module('oidc-media-config.js')).evaluate();
  await new Promise(resolve => setImmediate(resolve));
  return {elements, events, requests, save};
}

(async () => {
  let e = await mediaEnvironment();
  assert.equal(e.save.disabled, false);
  assert.equal(e.elements['oidc-media-aliases'].value, 'community.member');
  await e.events.submit({preventDefault() {}});
  const write = e.requests.find(r => r.options.method === 'PUT');
  assert.deepEqual(JSON.parse(write.options.body), {
    config: {clientIds: ['arbitrary-app'], scopeAliases: ['community.member']}, version: 7, reason: 'Test policy'
  });
  e = await mediaEnvironment(true);
  assert.equal(e.save.disabled, true);
  assert.equal(e.elements['oidc-media-status'].textContent, 'Load failed');
  assert.equal(e.requests.filter(r => r.options.method === 'PUT').length, 0);
  console.log('Passed: media policy load/save, arbitrary client, aliases, version/reason, failed-load save protection.');
})().catch(error => {console.error(error); process.exitCode = 1;});
