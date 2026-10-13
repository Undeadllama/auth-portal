# Registration-bound trusted OIDC consent

Depends on the strict configuration foundation; leaves upstream login, MFA and admission behavior intact.


In Admin → OAuth/OIDC → Trusted applications, select explicitly trusted clients
and save with a reason. The server snapshots the client's current registered
redirect URI and scope sets in the `oidc-consent` document. The PUT API accepts
client identifiers, not caller-supplied redirect/scope snapshots.

Trust bypasses ordinary first-use consent only while those registration sets still
match. Changes to registered redirects/scopes deactivate bypass until reviewed and
saved again. `prompt=consent` always requests consent. Unselected clients retain
the normal consent flow. Trust does not bypass login, MFA or mandatory media checks.


GET `/api/admin/oauth/consent` returns the persisted config and version. PUT takes `clientIds`, current `version` and a change `reason`; both routes require corresponding OAuth read/write permissions. Defaults are empty; trusted scopes/redirects are snapshotted from registration. Tests cover exact registration sets, changed registrations, validation and anonymous API denial.
