import {oidcConfigAPI as api} from './oidc-config-api.js';

const form = document.getElementById('oidc-consent-form');
if (form) {
  const list = document.getElementById('oidc-consent-clients');
  const status = document.getElementById('oidc-consent-status');
  let version = 0;

  const same = (a, b) => JSON.stringify([...new Set(a)].sort()) === JSON.stringify([...new Set(b)].sort());
  const save = form.querySelector('button[type="submit"]');
  async function load() {
    save.disabled = true;
    try {
      const [policy, clients] = await Promise.all([api('oauth/consent'), api('oauth/clients')]);
      version = policy.version; list.replaceChildren();
      clients.clients.forEach(client => {
        const trusted = policy.config.clients.find(p => p.clientId === client.clientId);
        const active = trusted && same(trusted.scopes, client.scopes) && same(trusted.redirectUris, client.redirectUris);
        const label = document.createElement('label');
        const checkbox = document.createElement('input'); checkbox.type = 'checkbox'; checkbox.value = client.clientId; checkbox.checked = !!active;
        label.append(checkbox, document.createTextNode(' ' + client.name + (trusted && !active ? ' (settings changed; trust inactive)' : '')));
        list.append(label, document.createElement('br'));
      });
      status.textContent = 'Trusted applications loaded.'; save.disabled = false;
    } catch (error) { status.textContent = error.message; }
  }
  document.getElementById('oidc-consent-reload').addEventListener('click', load);
  form.addEventListener('submit', async event => {
    event.preventDefault();
    try {
      const clientIds = Array.from(list.querySelectorAll('input:checked'), c => c.value);
      await api('oauth/consent', {clientIds, version, reason: document.getElementById('oidc-consent-reason').value});
      await load(); status.textContent = 'Trusted applications saved.';
    } catch (error) { status.textContent = error.message; }
  });
  load();
}
