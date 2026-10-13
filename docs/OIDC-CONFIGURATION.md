# OIDC policy configuration

All examples use fictitious identifiers. Register a client under Admin → OAuth/OIDC
first. Create custom roles/permissions under Access Control before referencing them.
There are no new application-specific environment variables or required scope names.
Policies are persisted in the existing versioned `app_config` store.

## Mandatory media admission

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

## Role and permission claims

In Access Control → OIDC role assignment, add a client policy and choose existing
roles for export. The `oidc-role-assignment` document is:

```json
{
  "clients": [{
    "clientId": "example-gallery-client",
    "exportRoles": ["gallery-basic", "gallery-library"],
    "exportPermissions": true
  }]
}
```

A configured client receives `roles` and `permissions` arrays in ID-token and
userinfo claims. Roles are the intersection of assigned roles and that client's
allowlist; permissions come only from exported roles when enabled. Empty results
are `[]`. Clients without a policy receive no additional role/permission claims.

Manual, LDAP and system bindings retain their existing assignment meaning.
Automatic `oidc-plex` bindings contribute claims only to their originating client.
Actual AuthPortal RBAC permissions remain global; claim filtering is not an
AuthPortal permission sandbox. Use dedicated downstream roles and avoid assigning
unwanted portal permissions.

## Optional Plex sharing rule

Add `plex` to the client policy above:

```json
{
  "ownerUuid": "plex-123456",
  "sectionKey": "42",
  "matchRole": "gallery-library",
  "defaultRole": "gallery-basic"
}
```

Use the verified existing Plex owner subject and exact library section key.
The owner is treated as a match. A verified share of the selected library uses
`matchRole`; a verified non-match uses `defaultRole`. Malformed/unavailable Plex
responses fail claim generation rather than granting a guessed role. Both roles
must exist, be custom, be exported by this client, differ from each other and
contain no built-in system permissions.

Assignments are tagged `source=oidc-plex`, `external_ref=<clientId>`. Changing,
disabling or removing a rule removes only that client's owned automatic bindings
in the same transaction as configuration/history. Unchanged rules and manual,
LDAP, system and other-client bindings are preserved. The next successful claim
generation under a rule computes the current assignment again.

## Trusted application consent

In Admin → OAuth/OIDC → Trusted applications, select explicitly trusted clients
and save with a reason. The server snapshots the client's current registered
redirect URI and scope sets in the `oidc-consent` document. The PUT API accepts
client identifiers, not caller-supplied redirect/scope snapshots.

Trust bypasses ordinary first-use consent only while those registration sets still
match. Changes to registered redirects/scopes deactivate bypass until reviewed and
saved again. `prompt=consent` always requests consent. Unselected clients retain
the normal consent flow. Trust does not bypass login, MFA or mandatory media checks.

## Admin API contract

| Resource | GET permission | PUT permission | PUT body |
|---|---|---|---|
| `/api/admin/oauth/media-authorization` | `oauth.read` | `oauth.write`; also `rbac.write` for alias changes | `{"config":{"clientIds":[],"scopeAliases":[]},"version":0,"reason":"Initial media policy"}` |
| `/api/admin/rbac/oidc-role-assignment` | `rbac.read` | `rbac.write` | `{"config":{"clients":[]},"version":0,"reason":"Initial role policy"}` |
| `/api/admin/oauth/consent` | `oauth.read` | `oauth.write` | `{"clientIds":[],"version":0,"reason":"Initial trust policy"}` |

GET returns an `ok`, `version` and `config` envelope. Read first and send the actual
version on PUT; version zero is only the absent document's initial version. Writes
use the existing same-origin guard and authenticated permission middleware.
Unknown fields, multiple JSON documents and requests above 64 KiB are rejected.
Reasons must be nonempty and at most 200 bytes after trimming. Stale versions,
including a concurrent first save, return HTTP 409. Reload and reconcile changes;
do not blindly retry by overwriting another administrator's policy. Failed initial
UI loads keep Save disabled.

## Login cancellation

An OIDC sign-in carries a local `/oidc/authorize?...` continuation. The page retains
it per tab and offers Cancel. Plex opens a popup only from a click; blocked popups
allow retry. Polling/MFA continuation is retained, and cancelled/closed flows suppress
late completion. Emby/Jellyfin flows also expose cancellation and clean up popup timers.

`POST /oidc/login/cancel` accepts form field `next`. It validates a local authorize
continuation, client registration and the exact registered callback before returning
JSON `{"redirect":"..."}` with `error=access_denied`, the original `state`, and no
code. Same-origin checks apply; no authenticated session is required to cancel an
unfinished login. Downstream applications may display their ordinary OAuth error
page. Application-specific friendly-error proxy redirects are outside this branch.
