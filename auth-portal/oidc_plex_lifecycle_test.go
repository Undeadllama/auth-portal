package main

import (
	"auth-portal/configstore"
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"testing"
)

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
