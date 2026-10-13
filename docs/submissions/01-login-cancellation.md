# Preserve OIDC login continuation and add safe cancellation

> **AI-generated code:** All code added or modified for this change was fully generated using AI. It has been tested, but must be independently reviewed by a maintainer before merging.

Target: `modom-ofn/auth-portal:main`. Prepared source: `submission/01-login-cancellation`.

A blocked popup or cancelled provider sign-in can leave an OIDC application without a usable continuation. This change opens Plex only from a user click, preserves per-tab continuation through polling/MFA, allows retry, and provides cancellation for Plex, Emby and Jellyfin. Cancellation validates the local authorize request and registered callback before returning access_denied with the original state; late callbacks and popup timers are cleaned up.

Configuration/defaults and implementation details: [OIDC-LOGIN-CANCELLATION.md](https://github.com/Undeadllama/auth-portal/blob/submission/01-login-cancellation/docs/OIDC-LOGIN-CANCELLATION.md).

Validation: Go callback/continuation tests, full upstream suite, Node popup/retry/cancel/MFA/provider regressions, vet and build. [Branch CI](https://github.com/Undeadllama/auth-portal/actions/runs/38418128673). The combined implementation was also deployed and operator-tested downstream; this split branch has separate local/CI validation and was not separately deployed over production.

Compatibility/limits: No policy configuration or credential endpoint is added. Friendly application-specific error redirects remain downstream responsibility.

GPL-3.0 licensing is retained.

Static analysis: continuation parsing now operates on query values without creating
an HTTP request; promises are explicitly handled and tests use a fixed repository
asset. The remaining Sonar finding concerns deliberate `vm.runInNewContext` use
in the browser regression harness. The source is the repository login asset, not
CLI-supplied code; browser APIs are mocked. Please independently review that test
execution and disposition the finding as appropriate. No security rule is suppressed.
The follow-up passed isolated Go/race/vet on the deployment host and Node regressions.
Production remains on the earlier validated combined image.
