package main

import (
	"reflect"
	"testing"
)

func TestAutomaticClaimsAreClientScoped(t *testing.T) {
	assignments := []RoleAssignment{{Role: "manual", Source: "manual"}, {Role: "ldap", Source: "ldap"}, {Role: "own", Source: "oidc-plex", ExternalRef: "a"}, {Role: "other", Source: "oidc-plex", ExternalRef: "b"}}
	if got := roleNamesForClient(assignments, "a"); !reflect.DeepEqual(got, []string{"ldap", "manual", "own"}) {
		t.Fatal(got)
	}
}

func TestEmptyRoleClaimsRemainArrays(t *testing.T) {
	roles, permissions := selectRoleClaims(nil, OIDCRolePolicy{ExportPermissions: true}, nil)
	if roles == nil || permissions == nil || len(roles) != 0 || len(permissions) != 0 {
		t.Fatal("empty role/permission claims must be arrays", roles, permissions)
	}
}
