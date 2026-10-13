# OIDC enhancement design and compatibility

## Admission, scopes and claims are separate decisions

Client selection expresses mandatory media admission. Requested scopes cannot
turn it off. The same policy is consulted during authorization, consent submission,
code exchange, refresh and userinfo, so ongoing requests revalidate media access.
Custom compatibility aliases preserve downstream scope names; they do not act as
application identifiers. Standard scopes retain upstream meaning.

Role exports are an additional per-client claim policy. Manual/LDAP/system roles
are filtered through the export allowlist. Generated bindings are also filtered by
originating client. The underlying RBAC model is deliberately unchanged: a role's
AuthPortal permissions are global even when its generated OIDC claims are filtered.

Trusted consent is an independent optional decision. Its saved redirect/scope sets
are compared as normalized sets against the registration. A mismatched registration
falls back to ordinary consent. Explicit consent remains authoritative.

## Configuration transactions and concurrency

The configstore receives optional `RequireVersion` and `BeforeCommit` update options.
Existing callers retain their previous behavior. New policy writes opt into version
comparison even when the expected version is zero and acquire a transaction-scoped
PostgreSQL advisory lock for the namespace/document. Configuration, history and any
dependent binding cleanup commit or roll back together.

Role cleanup compares previous and new Plex rules. It deletes `oidc-plex` records
for changed/removed client rules only. The assignment path obtains the same document
lock and rereads the persisted policy before inserting a binding. A login whose
rule was removed or changed while a network lookup was in progress is rejected.
The Plex HTTP request runs outside the transaction lock.

Plex requests use the fixed Plex service, a timeout, no redirects and a response-size
limit. The configured rule references an existing owner identity; it does not accept
an arbitrary endpoint or exported credential. Invalid identities, XML and shares are
rejected. Automatic role validation forbids system roles and system permissions.

## Defaults, rollout and rollback

No new SQL tables or required configuration values are introduced. The three policy
namespaces use existing configuration/history tables; assignment metadata uses
existing `user_roles` fields. Missing policy documents mean empty selections.
Unconfigured clients retain upstream admission, consent and role-claim behavior.
Login-page improvements apply independently of policy selections.

For stock upstream installations, configure only intended clients after deployment.
No automatic app-name detection or private migration is required. Downstream users
who already have mandatory guards must explicitly seed equivalent client selection
before replacing their custom image, or they would remove their admission requirement.
Preserve any scope aliases required by their integrations.

Back up the existing database and configuration before rollout. The old upstream
image ignores the new policy documents, but automatic assignments are real RBAC
records: image rollback alone does not undo them. Review/remove owned generated
bindings as appropriate without overwriting manual/LDAP/system assignments or
subsequent activity. Stopping role exports does not revoke already issued ID tokens
or downstream application sessions. This branch adds no downstream session-revocation
mechanism.

## Maintenance and scope boundaries

Shared server/browser helpers keep version handling, JSON validation, reasons and
failure messages consistent. Tests cover configuration conflicts, concurrent first
saves, transaction rollback, rule-removal races, exported-role isolation, empty-array
claims, provider failures, consent registration binding and browser continuation.

The prepared core source matches the deployed core patch. The live downstream image
also carries a separate pre-existing credential addon; that addon and its tests are
excluded here. Its endpoint is not a dependency of these enhancements. The public
branch is tested separately without it.

This branch does not change release tags, upstream Dockerfile/toolchain/dependencies,
registry publishing, connected-app provisioning, quotas, passwords or reverse proxies.
A focused contribution sequence and outstanding cross-platform checks are listed in
[OIDC-CONTRIBUTION-PLAN.md](OIDC-CONTRIBUTION-PLAN.md).
