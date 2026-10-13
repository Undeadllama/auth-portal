import {oidcConfigAPI as api} from './oidc-config-api.js';

const form = document.getElementById('oidc-media-form');
if (form) {
  const list = document.getElementById('oidc-media-clients');
  const aliases = document.getElementById('oidc-media-aliases');
  const status = document.getElementById('oidc-media-status');
  const save = form.querySelector('button[type="submit"]');
  let version = 0;
  async function load() {
    save.disabled = true;
    try {
      const [policy, clients] = await Promise.all([api('oauth/media-authorization'), api('oauth/clients')]);
      version = policy.version;
      list.replaceChildren();
      clients.clients.forEach(client => {
        const label = document.createElement('label');
        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox'; checkbox.value = client.clientId;
        checkbox.checked = policy.config.clientIds.includes(client.clientId);
        label.append(checkbox, document.createTextNode(' ' + client.name));
        list.append(label, document.createElement('br'));
      });
      aliases.value = policy.config.scopeAliases.join(', ');
      status.textContent = 'Media authorization settings loaded.';
      save.disabled = false;
    } catch (error) { status.textContent = error.message; }
  }
  document.getElementById('oidc-media-reload').addEventListener('click', load);
  form.addEventListener('submit', async event => {
    event.preventDefault(); save.disabled = true;
    try {
      await api('oauth/media-authorization', {
        config: {
          clientIds: Array.from(list.querySelectorAll('input:checked'), c => c.value),
          scopeAliases: aliases.value.split(',').map(v => v.trim()).filter(Boolean)
        },
        version, reason: document.getElementById('oidc-media-reason').value
      });
      await load();
    } catch (error) { status.textContent = error.message; save.disabled = false; }
  });
  load();
}
