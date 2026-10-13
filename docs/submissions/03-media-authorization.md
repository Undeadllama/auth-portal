# Configure mandatory media authorization per OIDC client

Target: `modom-ofn/auth-portal:main`. Prepared source: `submission/03-media-authorization`.

Open only after `submission/02-config-foundation` and its prerequisites merge into upstream `main`; rebase and rerun checks first. The current branch includes those prerequisite commits.

A requested scope cannot enforce mandatory media admission if a client omits that scope. Add an administrator-selected client policy that checks the active media provider at authorization/consent, code exchange, refresh and userinfo. Arbitrary registered applications use the same gate; optional custom aliases preserve existing scope names. OAuth managers cannot alter alias classification without RBAC write permission.

Configuration/defaults and implementation details: [OIDC-MEDIA-AUTHORIZATION.md](https://github.com/Undeadllama/auth-portal/blob/submission/03-media-authorization/docs/OIDC-MEDIA-AUTHORIZATION.md).

Validation: Full suite/race/PostgreSQL, arbitrary-client/provider-failure and omitted-scope authorization/code/userinfo checks, alias opt-out/permission checks, authenticated routes and Node media UI regressions. [Branch CI](https://github.com/Undeadllama/auth-portal/actions/runs/38408310519). The combined implementation was also deployed and operator-tested downstream; this split branch has separate local/CI validation and was not separately deployed over production.

Compatibility/limits: Empty selection preserves stock upstream admission. Deployments with existing custom mandatory guards must seed equivalent selections before replacing those guards. Fresh live admission on Emby/Jellyfin and mobile-specific acceptance remain useful additional coverage.

Development used AI assistance; code, automated results and deployment evidence were reviewed separately. GPL-3.0 licensing is retained.
