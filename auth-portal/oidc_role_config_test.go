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

func TestUnconfiguredClientHasNoRoleClaims(t *testing.T) {
	old := configStore
	configStore = nil
	defer func() { configStore = old }()
	claims, err := configuredOIDCRoleClaims("unconfigured", User{})
	if err != nil || claims != nil {
		t.Fatal(claims, err)
	}
}
