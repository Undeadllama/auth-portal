# Add opt-in strict OIDC configuration versions and transactional hooks

> **AI-generated code:** All code added or modified for this change was fully generated using AI. It has been tested, but must be independently reviewed by a maintainer before merging.

Target: `modom-ofn/auth-portal:main`. Prepared source: `submission/02-config-foundation`.

Configuration features need reliable concurrent initial saves and atomic dependent changes. Add optional RequireVersion and BeforeCommit configstore options, shared bounded/strict OIDC JSON and reason/version handling, and a shared browser API helper. Configuration/history/dependent updates share a PostgreSQL transaction and document advisory lock; existing callers keep their previous defaults.

Configuration/defaults and implementation details: [OIDC-CONFIGURATION-FOUNDATION.md](https://github.com/Undeadllama/auth-portal/blob/submission/02-config-foundation/docs/OIDC-CONFIGURATION-FOUNDATION.md).

Validation: Full upstream/race suite, real isolated PostgreSQL version-zero/concurrent creation/rollback checks, strict request tests and vet. [Branch CI](https://github.com/Undeadllama/auth-portal/actions/runs/38408306223). The combined implementation was also deployed and operator-tested downstream; this split branch has separate local/CI validation and was not separately deployed over production.

Compatibility/limits: This is infrastructure for the following OIDC policies. It enables no application policy or UI by itself and adds no tables.

GPL-3.0 licensing is retained.
