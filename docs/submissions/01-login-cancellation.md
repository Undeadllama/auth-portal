# Preserve OIDC login continuation and add safe cancellation

Target: `modom-ofn/auth-portal:main`. Prepared source: `submission/01-login-cancellation`.

A blocked popup or cancelled provider sign-in can leave an OIDC application without a usable continuation. This change opens Plex only from a user click, preserves per-tab continuation through polling/MFA, allows retry, and provides cancellation for Plex, Emby and Jellyfin. Cancellation validates the local authorize request and registered callback before returning access_denied with the original state; late callbacks and popup timers are cleaned up.

Configuration/defaults and implementation details: [OIDC-LOGIN-CANCELLATION.md](https://github.com/Undeadllama/auth-portal/blob/submission/01-login-cancellation/docs/OIDC-LOGIN-CANCELLATION.md).

Validation: Go callback/continuation tests, full upstream suite, Node popup/retry/cancel/MFA/provider regressions, vet and build. [Branch CI](https://github.com/Undeadllama/auth-portal/actions/runs/38408300234). The combined implementation was also deployed and operator-tested downstream; this split branch has separate local/CI validation and was not separately deployed over production.

Compatibility/limits: No policy configuration or credential endpoint is added. Friendly application-specific error redirects remain downstream responsibility.

Development used AI assistance; code, automated results and deployment evidence were reviewed separately. GPL-3.0 licensing is retained.
