# Configurable OIDC enhancements

This feature branch proposes six enhancements developed and tested in a downstream
AuthPortal deployment. It is based on upstream `main` commit
`62fea17bdda4085e6ee8b8588ed8f42af8e8463f` (AuthPortal v2.0.6 application code).
It is a contribution preparation branch, not an upstream release. No pull request
has been submitted. Existing upstream licensing remains GPL-3.0.

Start with [configuration and API examples](OIDC-CONFIGURATION.md),
[design and compatibility](OIDC-DESIGN.md), and
[validation and proposed submission sequence](OIDC-CONTRIBUTION-PLAN.md).
The [confidentiality audit](SECURITY-AUDIT.md) records the pre-submission review.

## Problems and resulting behavior

| Enhancement | Problem addressed | Result |
|---|---|---|
| Mandatory media authorization, originally used for Immich | A scope alone cannot require authorization when a caller omits it. | Administrators select any registered client that must have current access to the configured media provider. |
| Mandatory media authorization, originally used for Organizr | Application-specific guards require code edits for every new application. | The same client-selection policy covers arbitrary applications, with optional compatibility scope aliases. |
| Login continuation and cancellation | Popup blocking, incomplete callbacks and cancelled sign-ins can leave an OIDC flow stranded. | Click-only Plex popup startup, retry, per-tab continuation and safe registered-callback cancellation for Plex/Emby/Jellyfin. |
| Configurable role and permission claims | Downstream clients need different role exports without hardcoded application mappings. | Per-client export allowlists add `roles` and optional `permissions` to ID tokens/userinfo. |
| Automatic Plex library-sharing assignments | Shared library access should map to dedicated downstream roles without manual duplication. | Optional owner/library/match/default rules assign owned bindings, with atomic cleanup and race protection. |
| Configurable trusted consent | First-party clients may need first-use consent bypass without disabling consent globally. | Explicit admin trust is bound to each client's registered redirects/scopes; `prompt=consent` still works. |

## Safe defaults and boundaries

New policy documents start empty. No application is selected for mandatory media
admission, trusted consent or additional role claims automatically. There are no
private client identifiers, account mappings, domains or deployments in this branch.
The two original media guards become one configurable feature; no application-specific
runtime branches are introduced.

Media admission works with the existing active provider interface. Automatic
library-sharing assignment is explicitly Plex-specific; this proposal does not add
Emby/Jellyfin library-sharing adapters. Role/permission exports and trusted consent
are provider-independent.

This branch excludes the downstream confidential Plex credential endpoint and
all application bridges, proxy rules, image overlays and production migration
scripts. It adds no credential-export route. Administrators configure existing
clients and custom roles through the ordinary admin interface.

## Review map

- `auth-portal/oidc_media_config.go`: client selection and current provider authorization.
- `auth-portal/login_oauth.go`, `static/login.js`: validated cancellation and browser continuation.
- `auth-portal/oidc_role_config.go`: claim allowlists, Plex rules and owned binding lifecycle.
- `auth-portal/oidc_consent_config.go`: registration-bound trusted consent.
- `auth-portal/oidc_config_helpers.go`, `static/admin/oidc-config-api.js`: shared strict admin configuration handling.
- `auth-portal/configstore/store.go`: opt-in strict version checks and transactional dependent updates.
- `auth-portal/oidc_handlers.go`, `main.go`, `templates/admin.html`: OIDC and admin integration.
- Go tests alongside the implementation; Node regressions in `auth-portal/tests/`.

The implementation is prepared as one integration branch to preserve the tested
behavior. The [prepared branch map and PR drafts](submissions/README.md) provide smaller independently validated PRs
rather than asking a maintainer to accept the entire branch at once.

Development used AI assistance. Automated tests, live deployment checks and
operator acceptance are recorded separately; downstream test success does not
replace upstream maintainer review or imply upstream acceptance.
