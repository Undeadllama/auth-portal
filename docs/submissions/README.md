# Prepared AuthPortal submissions

Six focused branches and review-ready draft descriptions are prepared. No PR is open.
The target is upstream `main`; maintainer discussion is treated as complete.

| Branch | Prerequisite | CI | Draft description |
|---|---|---|---|
| [01-login-cancellation](https://github.com/Undeadllama/auth-portal/tree/submission/01-login-cancellation) | upstream main | [success](https://github.com/Undeadllama/auth-portal/actions/runs/38408300234) | [Preserve OIDC login continuation and add safe cancellation](01-login-cancellation.md) |
| [02-config-foundation](https://github.com/Undeadllama/auth-portal/tree/submission/02-config-foundation) | upstream main | [success](https://github.com/Undeadllama/auth-portal/actions/runs/38408306223) | [Add opt-in strict OIDC configuration versions and transactional hooks](02-config-foundation.md) |
| [03-media-authorization](https://github.com/Undeadllama/auth-portal/tree/submission/03-media-authorization) | submission/02-config-foundation | [success](https://github.com/Undeadllama/auth-portal/actions/runs/38408310519) | [Configure mandatory media authorization per OIDC client](03-media-authorization.md) |
| [04-role-claims](https://github.com/Undeadllama/auth-portal/tree/submission/04-role-claims) | submission/02-config-foundation | [success](https://github.com/Undeadllama/auth-portal/actions/runs/38408316965) | [Export allowlisted roles and permissions in OIDC claims](04-role-claims.md) |
| [05-plex-role-assignment](https://github.com/Undeadllama/auth-portal/tree/submission/05-plex-role-assignment) | submission/04-role-claims | [success](https://github.com/Undeadllama/auth-portal/actions/runs/38408322044) | [Assign owned OIDC roles from Plex library sharing](05-plex-role-assignment.md) |
| [06-trusted-consent](https://github.com/Undeadllama/auth-portal/tree/submission/06-trusted-consent) | submission/02-config-foundation | [success](https://github.com/Undeadllama/auth-portal/actions/runs/38408327375) | [Allow registration-bound trusted OIDC consent](06-trusted-consent.md) |

## Submission order and rebasing

Login cancellation and the configuration foundation can be submitted independently
against current `main`. After the foundation merges, rebase media authorization,
role claims and trusted consent onto updated upstream `main`. They are independent
of one another. Plex assignment waits for role claims as well. Do not open these
cumulative branches against `main` before prerequisites merge unless intentionally
requesting a stacked review.

Each feature has one consolidated commit relative to its stated prerequisite. The final
branches retain upstream main/next push triggers; validation was dispatched explicitly
on each feature ref. CI fixture/race/PostgreSQL/Node checks accompany the changes.
No fork-only submission trigger remains.

Upstream may squash or modify prerequisites. Record the prerequisite tip shown in
`manifest.json` rather than assuming its commit will appear unchanged in `main`.
For an unchanged feature branch, after fetching upstream:

```sh
# Example only, after the foundation is merged; not executed during preparation.
git switch submission/03-media-authorization
git rebase --onto upstream/main <recorded-foundation-tip> submission/03-media-authorization
```

Do not use a moving prerequisite reference once its history has changed. Handle
shared router, OIDC-handler and admin-template conflicts by retaining already merged
features and applying only the next feature's changes. Never overwrite these files
with an old combined snapshot after upstream has moved. Recheck the complete diff,
permissions/defaults and CI before opening the PR.

## Validation and environment boundaries

Every branch has its own full Go suite/vet and fork CI. Feature-specific browser
regressions run only where their controls exist; PostgreSQL tests use a disposable
service. Secret scanning is bounded to the contribution range (new branches fall
back to the merge base with fork main) and public fixture DSNs are constructed at
runtime. Intermediate preparation history was consolidated; locally retained refs
preserve the previous task commits. No secret detector is disabled.

A local reassembly of all six features passed Go/vet and upstream/new Node suites.
Its 15 deployment runtime source hashes and public main.go match the previously
validated combined implementation. Generic tests were redistributed by feature.
Shared integration conflicts were resolved deliberately; later rebases still need
review. The local assembly is not published or deployed.

Current production remains the tested combined image. No runtime settings, images,
accounts or policies were changed in this branch-preparation task. Live combined
browser/Plex evidence does not claim separate deployments of every split branch.
Fresh Emby/Jellyfin admission, representative LDAP, mobile-specific acceptance and
backup restoration remain useful documented coverage gaps, not reported successes.

No PRs, issues/comments, releases or registry images were created. Authorizing branch
preparation does not authorize submitting these PRs.
