import {oidcConfigAPI as api} from './oidc-config-api.js';

const form = document.getElementById('oidc-role-config-form');
if (form) {
  const el = (name) => document.getElementById('oidc-role-' + name);
  let config = {clients: []}, version = 0;

  function show() {
    const policy = config.clients.find(p => p.clientId === el('client').value) || {exportRoles: []};
    el('exports').value = (policy.exportRoles || []).join(', ');
    el('permissions').checked = !!policy.exportPermissions;
    el('plex').checked = !!policy.plex;
    el('owner').value = policy.plex?.ownerUuid || '';
    el('section').value = policy.plex?.sectionKey || '';
    el('match').value = policy.plex?.matchRole || '';
    el('default').value = policy.plex?.defaultRole || '';
  }
  const save = form.querySelector('button[type="submit"]');
  async function load() {
    save.disabled = true;
    try {
      const selected = el('client').value;
      const [settings, clients, rbac] = await Promise.all([api('rbac/oidc-role-assignment'), api('oauth/clients'), api('rbac')]);
      config = settings.config; version = settings.version;
      el('client').replaceChildren(new Option('Select application', ''));
      clients.clients.forEach(c => el('client').append(new Option(c.name, c.clientId)));
      el('client').value = selected || config.clients[0]?.clientId || '';
      for (const field of ['match', 'default']) {
        el(field).replaceChildren(new Option('Select role', ''));
        rbac.roles.filter(r => !r.system).forEach(r => el(field).append(new Option(r.name, r.name)));
      }
      show(); el('status').textContent = 'Configuration loaded.'; save.disabled = false;
    } catch (error) { el('status').textContent = error.message; }
  }
  el('client').addEventListener('change', show);
  el('reload').addEventListener('click', load);
  form.addEventListener('submit', async event => {
    event.preventDefault();
    const policy = {clientId: el('client').value, exportRoles: el('exports').value.split(',').map(v => v.trim()).filter(Boolean), exportPermissions: el('permissions').checked};
    if (el('plex').checked) policy.plex = {ownerUuid: el('owner').value.trim(), sectionKey: el('section').value.trim(), matchRole: el('match').value, defaultRole: el('default').value};
    const next = {clients: config.clients.filter(p => p.clientId !== policy.clientId).concat(policy)};
    try {
      await api('rbac/oidc-role-assignment', {config: next, version, reason: el('reason').value});
      await load(); el('status').textContent = 'Role assignment saved. Existing sessions are unchanged.';
    } catch (error) { el('status').textContent = error.message; }
  });
  load();
}
