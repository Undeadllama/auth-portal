# Plex sharing-based OIDC role assignment

Depends on per-client role claims and its configuration prerequisite; target upstream `main` after they merge.


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


Plex requests have a timeout, fixed endpoint, no redirects and bounded response size. Generated roles retain global AuthPortal RBAC semantics; client isolation applies to OIDC claims. No credential-export endpoint is added. Tests cover shares, system-permission denials, owned cleanup and removed-policy races, plus real PostgreSQL transactions in the foundation.
