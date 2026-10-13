package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"auth-portal/configstore"
	"auth-portal/oauth"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

func setTestOIDCConfig(t *testing.T, section configstore.Section, payload any) {
	t.Helper()
	testDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery("SELECT namespace, key, value").WillReturnRows(sqlmock.NewRows([]string{"namespace", "key", "value", "version", "updated_at", "updated_by"}).AddRow(string(section), configstore.SectionDocumentKey, raw, 1, time.Now(), "test"))
	store, err := configstore.New(testDB, configstore.Options{})
	if err != nil {
		t.Fatal(err)
	}
	old := configStore
	configStore = store
	t.Cleanup(func() {
		configStore = old
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		testDB.Close()
	})
}

func TestMediaPolicyIsClientConfiguredAndScopeIndependent(t *testing.T) {
	setTestOIDCConfig(t, oidcMediaSection, OIDCMediaConfig{ClientIDs: []string{"arbitrary-app"}})
	old := activeProvider()
	t.Cleanup(func() { setCurrentProvider(old) })
	for _, tc := range []struct {
		name    string
		allowed bool
		failure error
		denied  bool
	}{
		{"allowed", true, nil, false}, {"denied", false, nil, true}, {"offline", true, errors.New("offline"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider := &mediaTestProvider{allowed: tc.allowed, err: tc.failure}
			setCurrentProvider(provider)
			if err := enforceOIDCMediaPolicy("arbitrary-app", "plex-test", "test-user"); (err != nil) != tc.denied {
				t.Fatal(err)
			}
			if provider.calls != 1 {
				t.Fatal("provider not checked")
			}
			if err := enforceOIDCMediaPolicy("unselected", "plex-test", "test-user"); err != nil {
				t.Fatal(err)
			}
			if provider.calls != 1 {
				t.Fatal("unselected client checked")
			}
		})
	}
	setCurrentProvider(nil)
	if enforceOIDCMediaPolicy("arbitrary-app", "plex-test", "test-user") == nil {
		t.Fatal("missing provider admitted")
	}
}

func TestGenericMediaAliasesCannotBeOmitted(t *testing.T) {
	setTestOIDCConfig(t, oidcMediaSection, OIDCMediaConfig{ScopeAliases: []string{"community.member"}})
	requested := []string{"openid"}
	got, err := enforceClientScopePolicy(requested, oauth.Client{Scopes: []string{"openid", "community.member"}})
	if err != nil || !containsScope(got, "community.member") || len(requested) != 1 {
		t.Fatal(got, err, requested)
	}
}

func TestMediaConfigValidation(t *testing.T) {
	for _, cfg := range []OIDCMediaConfig{
		{ClientIDs: []string{"a", "a"}}, {ClientIDs: []string{" a"}},
		{ScopeAliases: []string{"openid"}}, {ScopeAliases: []string{"admin.access"}}, {ScopeAliases: []string{"a b"}}, {ScopeAliases: []string{"a", "a"}},
	} {
		if validateOIDCMediaConfig(cfg) == nil {
			t.Fatal("invalid policy accepted", cfg)
		}
	}
	if err := validateOIDCMediaConfig(OIDCMediaConfig{ClientIDs: []string{"any-app"}, ScopeAliases: []string{"community.member"}}); err != nil {
		t.Fatal(err)
	}
}

func TestInvalidMediaConfigurationFailsClosed(t *testing.T) {
	setTestOIDCConfig(t, oidcMediaSection, map[string]any{"clientIds": []string{"app", "app"}})
	if enforceOIDCMediaPolicy("app", "plex-test", "test-user") == nil {
		t.Fatal("invalid configuration admitted")
	}
	if _, err := enforceClientScopePolicy([]string{"openid"}, oauth.Client{}); err == nil {
		t.Fatal("scope policy ignored invalid configuration")
	}
}

func TestStrictConfigRequest(t *testing.T) {
	for _, body := range []string{`{"unknown":1}`, `{} {}`, strings.Repeat(" ", 65536) + `{}`, `{`} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", "/", strings.NewReader(body))
		var target struct {
			Version int `json:"version"`
		}
		if decodeOIDCConfigRequest(w, r, &target) || w.Code != 400 {
			t.Fatal("invalid request accepted", w.Code)
		}
	}
	for _, reason := range []string{"", strings.Repeat("x", 201)} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", "/", nil)
		if saveOIDCSection(w, r, oidcMediaSection, OIDCMediaConfig{}, 0, reason) || w.Code != 400 {
			t.Fatal("invalid reason accepted")
		}
	}
}

func TestConfigRoutesRequireAuthentication(t *testing.T) {
	router := mux.NewRouter()
	registerAdminRoutes(router)
	for _, path := range []string{"/api/admin/oauth/media-authorization", "/api/admin/oauth/consent", "/api/admin/rbac/oidc-role-assignment"} {
		for _, method := range []string{"GET", "PUT"} {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(method, path, strings.NewReader(`{}`))
			r.Header.Set("Accept", "application/json")
			router.ServeHTTP(w, r)
			if w.Code != 401 && w.Code != 403 && w.Code != 302 {
				t.Fatal(path, method, w.Code)
			}
		}
	}
}

func TestSelectedClientCannotAuthorizeWithOnlyOpenID(t *testing.T) {
	setTestOIDCConfig(t, oidcMediaSection, OIDCMediaConfig{ClientIDs: []string{"arbitrary-app"}})
	testDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	oldService, oldProvider := oauthService, activeProvider()
	oauthService = oauth.Service{DB: testDB}
	defer func() { oauthService = oldService; setCurrentProvider(oldProvider) }()
	setCurrentProvider(&mediaTestProvider{allowed: false})
	mock.ExpectQuery("SELECT client_id, client_secret").WithArgs("arbitrary-app").WillReturnRows(sqlmock.NewRows([]string{"client_id", "client_secret", "name", "redirect_uris", "scopes", "grant_types", "response_types", "created_at", "updated_at"}).AddRow("arbitrary-app", "", "Example", pq.StringArray{"https://app.example/callback"}, pq.StringArray{"openid"}, pq.StringArray{"authorization_code"}, pq.StringArray{"code"}, time.Now(), time.Now()))
	ctx := withUsername(withUUID(context.Background(), "plex-test"), "test-user")
	_, _, _, err = prepareAuthorizeClient(ctx, authorizeRequest{ClientID: "arbitrary-app", RedirectURI: "https://app.example/callback", Scopes: []string{"openid"}})
	if err == nil {
		t.Fatal("selected application authorized denied user without media scope")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSelectedClientUserinfoRechecksMediaAccessWithoutScope(t *testing.T) {
	setTestOIDCConfig(t, oidcMediaSection, OIDCMediaConfig{ClientIDs: []string{"a"}})
	testDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	oldDB, oldService, oldProvider := db, oauthService, activeProvider()
	db = testDB
	oauthService = oauth.Service{DB: testDB}
	defer func() { db = oldDB; oauthService = oldService; setCurrentProvider(oldProvider) }()
	setCurrentProvider(&mediaTestProvider{allowed: false})
	mock.ExpectQuery("SELECT token_id, client_id, user_id, scopes").WithArgs(sqlmock.AnyArg()).WillReturnRows(sqlmock.NewRows([]string{"token_id", "client_id", "user_id", "scopes", "expires_at", "revoked_at", "created_at"}).AddRow("digest", "a", 42, pq.StringArray{"openid"}, time.Now().Add(time.Hour), nil, time.Now()))
	mock.ExpectQuery("SELECT id, username, email").WithArgs(42).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "media_uuid", "media_token", "media_access", "is_admin", "admin_granted_at", "admin_granted_by", "session_version"}).AddRow(42, "test-user", "user@example.com", "plex-test", nil, false, false, nil, nil, 1))
	r := httptest.NewRequest("GET", "/oidc/userinfo", nil)
	r.Header.Set("Authorization", "Bearer test-access-token")
	w := httptest.NewRecorder()
	oidcUserinfoHandler(w, r)
	if w.Code != 403 {
		t.Fatal("userinfo did not recheck media access", w.Code, w.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAutomaticClaimsAreClientScoped(t *testing.T) {
	assignments := []RoleAssignment{{Role: "manual", Source: "manual"}, {Role: "ldap", Source: "ldap"}, {Role: "own", Source: "oidc-plex", ExternalRef: "a"}, {Role: "other", Source: "oidc-plex", ExternalRef: "b"}}
	if got := roleNamesForClient(assignments, "a"); !reflect.DeepEqual(got, []string{"ldap", "manual", "own"}) {
		t.Fatal(got)
	}
}

func TestPlexRuleLifecycle(t *testing.T) {
	rule := &PlexRoleRule{OwnerUUID: "plex-1", SectionKey: "23", MatchRole: "anime", DefaultRole: "media"}
	old := OIDCRoleConfig{Clients: []OIDCRolePolicy{{ClientID: "a", Plex: rule}}}
	if len(changedPlexPolicies(old, old)) != 0 {
		t.Fatal("unchanged rule pruned")
	}
	for _, next := range []OIDCRoleConfig{{}, {Clients: []OIDCRolePolicy{{ClientID: "a"}}}, {Clients: []OIDCRolePolicy{{ClientID: "a", Plex: &PlexRoleRule{SectionKey: "24"}}}}} {
		if got := changedPlexPolicies(old, next); !reflect.DeepEqual(got, []string{"a"}) {
			t.Fatal(got)
		}
	}
}

func TestRoleCleanupDeletesOnlyOwnedBindings(t *testing.T) {
	testDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	mock.ExpectBegin()
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectExec("DELETE FROM user_roles WHERE source='oidc-plex' AND external_ref=\\$1").WithArgs("a").WillReturnResult(sqlmock.NewResult(0, 3))
	old := OIDCRoleConfig{Clients: []OIDCRolePolicy{{ClientID: "a", Plex: &PlexRoleRule{}}}}
	if err := pruneOIDCPlexRoles(context.Background(), tx, old, OIDCRoleConfig{}); err != nil {
		t.Fatal(err)
	}
	mock.ExpectCommit()
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRemovedRuleCannotRecreateBindings(t *testing.T) {
	testDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	oldDB := db
	db = testDB
	defer func() { db = oldDB }()
	mock.ExpectBegin()
	mock.ExpectExec("SELECT pg_advisory_xact_lock").WithArgs(string(oidcRoleSection), configstore.SectionDocumentKey).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT value FROM app_config").WithArgs(string(oidcRoleSection), configstore.SectionDocumentKey).WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow([]byte(`{"clients":[]}`)))
	mock.ExpectRollback()
	if assignOIDCPlexRole(42, "a", "media", PlexRoleRule{MatchRole: "media"}) == nil {
		t.Fatal("removed rule recreated binding")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestMediaGateBlocksAuthorizationCodeExchangeWithoutScope(t *testing.T) {
	setTestOIDCConfig(t, oidcMediaSection, OIDCMediaConfig{ClientIDs: []string{"a"}})
	testDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	oldDB, oldProvider := db, activeProvider()
	db = testDB
	defer func() { db = oldDB; setCurrentProvider(oldProvider) }()
	setCurrentProvider(&mediaTestProvider{allowed: false})
	mock.ExpectQuery("SELECT id, username, email").WithArgs(42).WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "media_uuid", "media_token", "media_access", "is_admin", "admin_granted_at", "admin_granted_by", "session_version"}).AddRow(42, "test-user", "user@example.com", "plex-test", nil, false, false, nil, nil, 1))
	_, _, _, _, err = issueAuthCodeGrantTokens(context.Background(), oauth.Client{ClientID: "a"}, oauth.AuthCode{UserID: 42, Scopes: []string{"openid"}})
	if err == nil {
		t.Fatal("code exchange admitted denied identity")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type mediaTestProvider struct {
	allowed bool
	err     error
	calls   int
}

func (p *mediaTestProvider) Name() string                                { return "plex" }
func (p *mediaTestProvider) StartWeb(http.ResponseWriter, *http.Request) {}
func (p *mediaTestProvider) Forward(http.ResponseWriter, *http.Request)  {}
func (p *mediaTestProvider) IsAuthorized(uuid, username string) (bool, error) {
	p.calls++
	if uuid != "plex-test" || username != "test-user" {
		return false, errors.New("wrong identity")
	}
	return p.allowed, p.err
}

func TestEmptyRoleClaimsRemainArrays(t *testing.T) {
	roles, permissions := selectRoleClaims(nil, OIDCRolePolicy{ExportPermissions: true}, nil)
	if roles == nil || permissions == nil || len(roles) != 0 || len(permissions) != 0 {
		t.Fatal("empty role/permission claims must be arrays", roles, permissions)
	}
}

func TestCompatibilityAliasDoesNotOverrideClientSelection(t *testing.T) {
	setTestOIDCConfig(t, oidcMediaSection, OIDCMediaConfig{ClientIDs: []string{}, ScopeAliases: []string{"community.member"}})
	testDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	oldDB, oldProvider := db, activeProvider()
	db = testDB
	defer func() { db = oldDB; setCurrentProvider(oldProvider) }()
	provider := &mediaTestProvider{allowed: false}
	setCurrentProvider(provider)
	mock.ExpectQuery("SELECT u.id").WithArgs("plex-test").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectQuery("SELECT DISTINCT p.name").WithArgs(42).WillReturnRows(sqlmock.NewRows([]string{"name"}))
	if err := enforceOIDCMediaPolicy("deselected", "plex-test", "test-user"); err != nil {
		t.Fatal(err)
	}
	got, err := enforceOIDCUserScopeEntitlementsForIdentity("plex-test", "test-user", []string{"openid", "community.member"})
	if err != nil || !containsScope(got, "community.member") || provider.calls != 0 {
		t.Fatal("alias overrode client selection", got, err, provider.calls)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAliasClassificationRequiresRBACWrite(t *testing.T) {
	setTestOIDCConfig(t, oidcMediaSection, OIDCMediaConfig{ClientIDs: []string{}, ScopeAliases: []string{}})
	testDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	oldDB := db
	db = testDB
	defer func() { db = oldDB }()
	mock.ExpectQuery("SELECT u.id").WithArgs("plex-test").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	mock.ExpectQuery("SELECT EXISTS").WithArgs(42, sqlmock.AnyArg(), permissionAdminAccess).WillReturnRows(sqlmock.NewRows([]string{"matched", "admin_access"}).AddRow(false, false))
	body := `{"config":{"clientIds":[],"scopeAliases":["community.member"]},"version":1,"reason":"Attempt alias classification"}`
	r := httptest.NewRequest("PUT", "/api/admin/oauth/media-authorization", strings.NewReader(body))
	r = r.WithContext(withUsername(withUUID(r.Context(), "plex-test"), "test-user"))
	w := httptest.NewRecorder()
	adminOIDCMediaConfigHandler(w, r)
	if w.Code != 403 {
		t.Fatal("OAuth manager changed RBAC scope classification", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
