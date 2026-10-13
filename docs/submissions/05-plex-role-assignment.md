# Assign owned OIDC roles from Plex library sharing

Target: `modom-ofn/auth-portal:main`. Prepared source: `submission/05-plex-role-assignment`.

Open only after `submission/04-role-claims` and its prerequisites merge into upstream `main`; rebase and rerun checks first. The current branch includes those prerequisite commits.

Maintaining downstream roles manually duplicates Plex library-sharing state. Extend the client role policy with a verified owner/library/match/default rule. Only dedicated custom roles without system permissions may be assigned. Changing/removing a rule atomically cleans up that client's owned automatic bindings with configuration/history; assignment rereads current policy under the same lock to prevent stale login results recreating removed grants.

Configuration/defaults and implementation details: [OIDC-PLEX-ROLE-ASSIGNMENT.md](https://github.com/Undeadllama/auth-portal/blob/submission/05-plex-role-assignment/docs/OIDC-PLEX-ROLE-ASSIGNMENT.md).

Validation: Full suite/race/PostgreSQL, share/identity/system-permission validation, scoped binding cleanup, removed-rule race checks and Plex rule UI load/save regressions. [Branch CI](https://github.com/Undeadllama/auth-portal/actions/runs/38408322044). The combined implementation was also deployed and operator-tested downstream; this split branch has separate local/CI validation and was not separately deployed over production.

Compatibility/limits: Depends on role claims and the configuration foundation. Network lookup is bounded and outside the transaction lock. Generated roles grant real global AuthPortal RBAC permissions; dedicated downstream roles are recommended. Existing ID tokens/downstream sessions cannot be recalled by removing a rule. No credential-export route is added.

Development used AI assistance; code, automated results and deployment evidence were reviewed separately. GPL-3.0 licensing is retained.
