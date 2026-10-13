# Allow registration-bound trusted OIDC consent

Target: `modom-ofn/auth-portal:main`. Prepared source: `submission/06-trusted-consent`.

Open only after `submission/02-config-foundation` and its prerequisites merge into upstream `main`; rebase and rerun checks first. The current branch includes those prerequisite commits.

First-party applications may need first-use consent bypass without changing consent globally. Add explicit administrator trust for registered clients, bound to their current redirect and scope sets. Registration changes deactivate bypass until reviewed/saved again; prompt=consent still requests consent and unselected clients retain the upstream flow.

Configuration/defaults and implementation details: [OIDC-TRUSTED-CONSENT.md](https://github.com/Undeadllama/auth-portal/blob/submission/06-trusted-consent/docs/OIDC-TRUSTED-CONSENT.md).

Validation: Full suite/race/PostgreSQL, exact registration-set/change validation, explicit-consent behavior, anonymous API denial, and UI load/save/inactive-trust/error regressions. [Branch CI](https://github.com/Undeadllama/auth-portal/actions/runs/38408327375). The combined implementation was also deployed and operator-tested downstream; this split branch has separate local/CI validation and was not separately deployed over production.

Compatibility/limits: Trust starts empty and does not bypass authentication, MFA or authorization. PUT resolves registrations on the server rather than accepting caller-supplied trusted redirect/scope snapshots.

Development used AI assistance; code, automated results and deployment evidence were reviewed separately. GPL-3.0 licensing is retained.
