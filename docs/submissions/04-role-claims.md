# Export allowlisted roles and permissions in OIDC claims

Target: `modom-ofn/auth-portal:main`. Prepared source: `submission/04-role-claims`.

Open only after `submission/02-config-foundation` and its prerequisites merge into upstream `main`; rebase and rerun checks first. The current branch includes those prerequisite commits.

Different downstream clients need different claims without hardcoded application mappings. Add per-client role export allowlists and optional permissions derived from exported roles. ID tokens and userinfo gain claims only for configured clients, with empty results represented as arrays. Existing generated bindings are filtered by originating client while manual/LDAP/system assignment meaning remains unchanged.

Configuration/defaults and implementation details: [OIDC-ROLE-CLAIMS.md](https://github.com/Undeadllama/auth-portal/blob/submission/04-role-claims/docs/OIDC-ROLE-CLAIMS.md).

Validation: Full suite/race/PostgreSQL, role/permission allowlist and empty-array tests, origin filtering, unconfigured-client compatibility, anonymous API denial and standalone role UI load/save/error regressions. [Branch CI](https://github.com/Undeadllama/auth-portal/actions/runs/38408316965). The combined implementation was also deployed and operator-tested downstream; this split branch has separate local/CI validation and was not separately deployed over production.

Compatibility/limits: No automatic Plex rule is introduced here. Actual AuthPortal RBAC permissions remain global; claim filtering is not a portal permission sandbox. Representative live LDAP coverage remains useful.

Development used AI assistance; code, automated results and deployment evidence were reviewed separately. GPL-3.0 licensing is retained.
