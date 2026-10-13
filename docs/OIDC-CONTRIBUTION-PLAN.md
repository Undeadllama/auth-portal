# Validation and proposed contribution sequence

## Status and provenance

This preparation branch targets upstream `main` at
`62fea17bdda4085e6ee8b8588ed8f42af8e8463f`. Its application code matches v2.0.6
commit `c415c9f72184ec94074e2e6f6c2e4c945e18f568`; the intervening upstream changes
are release workflow/documentation changes. The six enhancements were developed
with AI assistance, reviewed and refactored into generic configuration, then built
and deployed in a downstream environment on 13 October 2026.

The downstream image carries an existing credential addon. That addon, private
configuration, account exports and infrastructure are excluded from this fork.
The core implementation is reproduced from the tested patch. The fork's Go and
Node suites are rerun against standalone upstream source without the addon.
No upstream PR, issue/comment, release or container publication has been created.

## Evidence and limits

| Layer | Evidence | Interpretation |
|---|---|---|
| Downstream build | Full upstream/custom Go suite with race detector; `go vet`; clean source reproduction | Tested generic implementation and original addon together. |
| Isolated PostgreSQL | Version-zero conflict, concurrent first-save, config/history/cleanup rollback and owned binding removal | Uses disposable fixture schemas/data; no production fixtures. |
| Browser regression harnesses | Plex click/retry/poll/MFA/late completion; all-provider cancellation/closure; admin policy load/save/error handling | Simulated browser behavior, not real provider UI automation. |
| Isolated candidate HTTP | Admin permissions/CSRF/version checks, signed client-scoped claims, consent, omitted-scope denials, cleanup and restart persistence | Runs the actual image with disposable fixtures. |
| Live downstream OIDC | Five registered applications passed current Plex admission; omitted-scope authorization/code/refresh/userinfo denial; RSA-verified ID tokens/userinfo; explicit consent and safe cancellation | Protocol tests used short-lived maintenance sessions; generated codes/tokens cleaned up. |
| Live compatibility | Existing accounts, registrations, roles/permissions and role/consent policies preserved; restart persistence verified; ordinary/owner role claims checked | Does not prove behavior on every AuthPortal deployment. |
| Operator acceptance | Successful fresh browser sign-in including MFA/Cancel, admin media-selection controls and connected-application login | Reported by the downstream operator separately from automated evidence. |
| Standalone fork | Go suite, vet, module/format checks and upstream/new Node smoke tests; fork CI results linked below | Excludes the credential addon. |

Mobile-specific acceptance and backup restoration were not separately confirmed.
Live media admission/automatic assignments were tested with Plex. Emby/Jellyfin
cancellation is covered by regression harnesses; fresh live admission against those
provider servers remains a useful cross-provider check. LDAP/system/manual claim
filtering has automated coverage; representative LDAP deployments would strengthen
confidence. No claim is made that upstream security scans or maintainer review have
already passed.

## Fork CI result

[CI run 38404476474](https://github.com/Undeadllama/auth-portal/actions/runs/38404476474)
passed on implementation/CI commit `b4c3e587720937ecec1520f9fd3fdb45f7f7a13e`:
Go module/format/vet checks and full race suite with isolated PostgreSQL;
upstream admin smoke plus both new Node regression suites; TruffleHog secret scan;
and the unmodified upstream Dockerfile container build (without publishing).
A documentation-only follow-up records this result; it does not change that tested
implementation. This CI does not include every upstream release/security workflow.

## Reproduce locally

From the repository root:

```sh
cd auth-portal
# Deterministic public fixture key/secret: never use these in deployment.
export DATA_KEY=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
export SESSION_SECRET=auth-portal-tests-only-01234567890123456789
go mod tidy -diff
go vet ./...
go test -race ./...
node static/admin/smoke-test.mjs
node tests/login_flow_test.cjs static/login.js
node --experimental-vm-modules tests/oidc_config_ui_test.cjs .
```

Use Go 1.26.8 and Node 24 to match upstream CI. Race tests require a supported C
compiler/CGO toolchain. PostgreSQL integration is skipped unless
`AUTHPORTAL_TEST_DATABASE_URL` is set. For that test, provide a dedicated disposable
PostgreSQL database with CREATE/DROP SCHEMA permission, then run:

```sh
go test -race ./configstore -run TestPostgresPolicyTransactions -count=1 -v
```

The test creates/drops its own unique schema, but should still never be pointed at
production. The fork CI supplies a disposable PostgreSQL service, exercises the
full race suite and runs both new Node regression scripts. Existing container-build
CI remains build-only. No release or registry workflow is part of validation.

## Proposed PR sequence

Do not submit this combined preparation branch wholesale without discussing scope.
The following is a proposal, not an agreed maintainer requirement:

1. **Login continuation/cancellation.** Independent change with Go callback-validation
   and Node popup/provider tests. Explain blocked-popup retry, MFA continuation and
   cancellation state preservation with before/after behavior.
2. **Strict versioned OIDC configuration foundation.** Optional configstore strict
   version/transaction hook plus shared helpers, with real PostgreSQL tests. Could
   be folded into the first configuration feature if preferred.
3. **Configurable media admission.** One generic feature replaces both originally
   application-specific guards. Include opt-in defaults, aliases, delegated-admin
   permissions, omitted-scope and provider-failure tests.
4. **Per-client role/permission claims.** Export allowlists and originating-client
   filtering; clearly distinguish OIDC claim isolation from global AuthPortal RBAC.
5. **Plex sharing role assignment/lifecycle.** Add generic owner/library rules on
   top of role export; include owned cleanup, concurrency, permission restrictions
   and network-failure behavior. Split or combine with role exports by agreement.
6. **Registration-bound trusted consent.** Opt-in trust and explicit consent,
   registration-change invalidation and shared configuration handling.

Shared changes to the handler/router/admin template must be reapplied carefully in
that order. Each resulting branch needs an independently passing suite and usable
admin flow; a combined green branch is not evidence that every future split passes.
Documentation should travel with each feature, and broad preparation notes can stay
in the fork rather than being included in every PR.

## Before opening upstream PRs

- Refresh upstream and check related issues/PRs again. At preparation time, open
  upstream PRs were dependency/action updates, not competing OIDC implementations.
- Confirm the preferred base (`main` versus `next`) and whether the maintainer wants
  these features, smaller PRs or an initial design discussion. The current README
  welcomes issues/PRs but does not prescribe CLA/DCO or a specific PR template;
  do not infer additional requirements without checking again.
- Perform the independent split above and review each full diff, API surface,
  UI wording, defaults and permissions. Keep unrelated formatting/dependencies out.
- Verify clean CI on each final commit: module consistency, formatting, vet,
  race/PostgreSQL tests, Node regressions, secret scan and upstream container build.
  Review any fork-only infrastructure failures separately from application failures.
- Where practical, run fresh Emby/Jellyfin admission and LDAP/manual/system role
  checks; repeat mobile and restoration tests if those become acceptance criteria.
- Review licensing/attribution and contribution rules. Keep GPL-3.0 intact; document
  AI assistance and actual evidence honestly. Check added code/history for secrets
  and deployment-specific values before every public push.
- Prepare concise PR titles/bodies with concrete problem/result, compatibility,
  meaningful test evidence, limitations and screenshots from fictitious data if
  requested. Do not post private deployment URLs or account records.
- Agree on version/release-note treatment with the maintainer; do not create a tag,
  publish an image or promise availability before acceptance.

No permission to open upstream PRs is implied by this preparation work.
