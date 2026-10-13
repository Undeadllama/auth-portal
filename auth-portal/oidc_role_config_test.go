package main

import (
	"reflect"
	"testing"
)

func TestConfiguredRoleClaims(t *testing.T) {
	p := OIDCRolePolicy{ExportRoles: []string{"custom-media"}, ExportPermissions: true}
	defs := []RoleDefinition{{Name: "custom-media", Permissions: []string{"media.read"}}, {Name: "admin", Permissions: []string{"admin.access"}}}
	roles, permissions := selectRoleClaims([]string{"admin", "custom-media"}, p, defs)
	if !reflect.DeepEqual(roles, []string{"custom-media"}) || !reflect.DeepEqual(permissions, []string{"media.read"}) {
		t.Fatal(roles, permissions)
	}
}
func TestConfigurablePlexRoleMatching(t *testing.T) {
	body := []byte(`<MediaContainer><SharedServer userID="42"><Section key="23" shared="1"/></SharedServer><SharedServer userID="43"><Section key="23" shared="0"/></SharedServer></MediaContainer>`)
	for _, tc := range []struct {
		id      string
		matched bool
	}{{"plex-42", true}, {"plex-43", false}} {
		match, err := plexShareMatches(body, tc.id, "plex-1", "23")
		if err != nil || match != tc.matched {
			t.Fatal(match, err)
		}
	}
	for _, bad := range []string{"plex-99", "user-42", "plex-0"} {
		if _, err := plexShareMatches(body, bad, "plex-1", "23"); err == nil {
			t.Fatal("unknown identity admitted")
		}
	}
	if _, err := plexShareMatches([]byte(`<MediaContainer><SharedServer userID="42"/><SharedServer userID="42"/></MediaContainer>`), "plex-42", "plex-1", "23"); err == nil {
		t.Fatal("duplicate identity")
	}
	if _, err := plexShareMatches([]byte(`<Other/>`), "plex-42", "plex-1", "23"); err == nil {
		t.Fatal("invalid root")
	}
}
func TestAutomaticRoleValidation(t *testing.T) {
	defs := []RoleDefinition{{Name: "media", Permissions: []string{"media.read"}}, {Name: "anime", Permissions: []string{"media.anime"}}, {Name: "danger", Permissions: []string{"admin.access"}}}
	p := OIDCRolePolicy{ClientID: "client", ExportRoles: []string{"media", "anime"}, Plex: &PlexRoleRule{OwnerUUID: "plex-1", SectionKey: "23", MatchRole: "anime", DefaultRole: "media"}}
	if err := validateOIDCRoleConfig(OIDCRoleConfig{Clients: []OIDCRolePolicy{p}}, defs); err != nil {
		t.Fatal(err)
	}
	p.Plex.MatchRole = "danger"
	p.ExportRoles = append(p.ExportRoles, "danger")
	if validateOIDCRoleConfig(OIDCRoleConfig{Clients: []OIDCRolePolicy{p}}, defs) == nil {
		t.Fatal("automatic system permission grant accepted")
	}
	p.Plex.MatchRole = "missing"
	if validateOIDCRoleConfig(OIDCRoleConfig{Clients: []OIDCRolePolicy{p}}, defs) == nil {
		t.Fatal("undefined role accepted")
	}
}
func TestUnconfiguredClientHasNoRoleClaims(t *testing.T) {
	old := configStore
	configStore = nil
	defer func() { configStore = old }()
	claims, err := configuredOIDCRoleClaims("unconfigured", User{})
	if err != nil || claims != nil {
		t.Fatal(claims, err)
	}
}
