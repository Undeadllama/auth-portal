# Configurable mandatory media authorization

Depends on the strict configuration foundation; target upstream `main` after that prerequisite merges.


In Admin → OAuth/OIDC → Media authorization, select the registered applications
that must pass the active media provider's `IsAuthorized` check. Save with a change
reason. Selection enforces admission regardless of the requested scopes, including
requests containing only `openid`.

The gate is checked at authorization/consent, authorization-code exchange, refresh
and userinfo. Provider errors or unavailable authorization deny selected clients.
Authorization returns `access_denied` to a validated registered callback; code and
refresh return `invalid_grant`; userinfo returns HTTP 403 on media denial.

The `oidc-media-authorization` document has this shape:

```json
{
  "clientIds": ["example-gallery-client"],
  "scopeAliases": ["community.member"]
}
```

Aliases are optional compatibility classification for existing custom scopes.
Clients registered with an alias retain that scope in issued tokens even when
it is omitted in an authorization request. An alias does not select an application
for media admission: the client checkbox controls that gate. Leave aliases empty
for new integrations that do not require historical scope names.

Aliases cannot be standard OIDC scopes or built-in system permissions. They classify
custom scopes as admitted compatibility scopes instead of ordinary RBAC scopes;
changing them therefore requires both `oauth.write` and `rbac.write`. Changing
only client selection requires `oauth.write`. Review alias changes as permission
changes and preserve aliases that existing integrations still need.


GET/PUT `/api/admin/oauth/media-authorization` use OAuth read/write permissions, current version, `config` and a nonempty `reason`; alias changes also require RBAC write. Empty selections preserve upstream defaults. See Go omitted-scope/provider/alias/admin regression tests and the Node media configuration test.
