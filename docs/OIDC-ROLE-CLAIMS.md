# Per-client OIDC role and permission claims

Depends on the strict configuration foundation. No Plex assignment rules are exposed by this branch.


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


GET/PUT `/api/admin/rbac/oidc-role-assignment` require RBAC read/write and writes use current `version`, `config` and `reason`. Go tests verify allowlists, unconfigured clients, origin filtering and array types.
