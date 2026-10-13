# Add opt-in strict OIDC configuration versions and transactional hooks

> **AI-generated code:** All code added or modified for this change was fully generated using AI. It has been tested, but must be independently reviewed by a maintainer before merging.

This PR provides shared configuration handling for four prepared OIDC enhancements: administrator-selected media authorization, per-client role/permission claims, Plex sharing-based role assignment, and trusted application consent. Their branches, feature documentation and draft descriptions are available in the [submission index](https://github.com/Undeadllama/auth-portal/blob/enhancement/generic-oidc-policies/docs/submissions/README.md).

### Why this foundation is needed

The new admin policies need reliable concurrent saves, consistent request validation and, for automatic role assignment, atomic changes to configuration and dependent bindings:

- **Concurrent first saves:** two administrators can both load an absent policy at version zero. After one creates version one, the other's version-zero save must return a conflict instead of overwriting it. Optional `RequireVersion` enables that comparison and serializes writes to the same policy document with a PostgreSQL transaction-scoped advisory lock.
- **Atomic role-policy changes:** changing or removing a Plex assignment rule must also remove that client's owned automatic role bindings. Optional `BeforeCommit` runs this cleanup in the configuration/history transaction, so the policy, audit entry and dependent changes commit or roll back together. The follow-up assignment path uses the same document lock and rereads the current rule before inserting a binding.
- **Consistent admin handling:** shared server helpers enforce bounded strict JSON, required change reasons and optimistic versions. The shared browser API helper provides consistent authenticated same-origin requests and error handling for the policy forms.

### Prepared follow-up contributions

| Feature | Administrator-visible behavior | Use of this foundation |
|---|---|---|
| [Media authorization](https://github.com/Undeadllama/auth-portal/tree/submission/03-media-authorization) | Select any registered application that must pass current media-provider authorization, including requests omitting custom media scopes. | Versioned client/alias policy and shared admin request/UI handling. |
| [Role and permission claims](https://github.com/Undeadllama/auth-portal/tree/submission/04-role-claims) | Choose the roles and optional permissions exported to each client in ID tokens/userinfo. | Versioned export allowlists and shared admin request/UI handling. |
| [Plex role assignment](https://github.com/Undeadllama/auth-portal/tree/submission/05-plex-role-assignment) | Assign dedicated roles from library-sharing rules; remove owned automatic bindings when a rule changes or is disabled. | Atomic cleanup with configuration/history, plus the shared document-lock convention. Also depends on role claims. |
| [Trusted consent](https://github.com/Undeadllama/auth-portal/tree/submission/06-trusted-consent) | Explicitly trust selected clients for first-use consent bypass, bound to registered scopes/redirects, while honoring explicit consent. | Versioned trust selection and shared admin request/UI handling. |

Each follow-up has separate validation and is intended for separate maintainer review. The dependent PRs will be opened against `main` after this foundation is accepted and merged, with branches rebased and checks rerun. The prepared branches are linked now so reviewers can assess the consumers and whether these abstractions are appropriate. Reviewing or merging this PR does not imply approval of those features. The independent [login cancellation PR #177](https://github.com/modom-ofn/auth-portal/pull/177) does not depend on this foundation.

### Compatibility and validation

The new configstore options are opt-in; existing callers retain their current defaults. This PR enables no new application policy or admin form and adds no database tables. [Foundation implementation/documentation](https://github.com/Undeadllama/auth-portal/blob/submission/02-config-foundation/docs/OIDC-CONFIGURATION-FOUNDATION.md).

[Branch CI passed](https://github.com/Undeadllama/auth-portal/actions/runs/38408306223): full Go/race suite, real isolated PostgreSQL version-zero/concurrent creation/rollback tests, strict-request tests, vet, existing admin smoke tests, secret scan and container build. The earlier combined implementation was deployed and operator-tested downstream; this split branch has separate local/CI validation and was not separately deployed over production. Upstream Sonar analysis passed; upstream Actions awaits maintainer approval for the fork contribution.

GPL-3.0 licensing is retained.
