package main

import (
	"auth-portal/oauth"
	"testing"
)

func TestTrustedConsentClientRegistrationBound(t *testing.T) {
	c := oauth.Client{ClientID: "trusted", RedirectURIs: []string{"https://app.example/callback"}, Scopes: []string{"openid", "immich.access"}}
	cfg := OIDCConsentConfig{Clients: []TrustedOIDCClient{{c.ClientID, c.RedirectURIs, c.Scopes}}}
	if !trustedConsentClient(cfg, c) {
		t.Fatal("configured registration not trusted")
	}
	c.ClientID = "other"
	if trustedConsentClient(cfg, c) {
		t.Fatal("other client trusted")
	}
	c.ClientID = "trusted"
	c.Scopes = []string{"openid", "immich.access", "email"}
	if trustedConsentClient(cfg, c) {
		t.Fatal("scope expansion inherited trust")
	}
	c.Scopes = []string{"immich.access", "openid"}
	if !trustedConsentClient(cfg, c) {
		t.Fatal("scope order changed trust")
	}
	c.RedirectURIs = []string{"https://other.example/callback"}
	if trustedConsentClient(cfg, c) {
		t.Fatal("redirect change inherited trust")
	}
	if trustedConsentClient(OIDCConsentConfig{}, c) {
		t.Fatal("empty policy trusted")
	}
}
func TestConsentExplicitPromptAlwaysWins(t *testing.T) {
	if !shouldRequireConsent(nil, User{}, oauth.Client{}, nil, true) {
		t.Fatal("explicit consent bypassed")
	}
}
func TestConsentConfigRejectsInvalidSelections(t *testing.T) {
	for _, ids := range [][]string{{""}, {"a", "a"}, make([]string, 101)} {
		if validateConsentClients(ids) == nil {
			t.Fatal("invalid selection accepted")
		}
	}
	if validateConsentClients([]string{"a", "b"}) != nil {
		t.Fatal("valid selection rejected")
	}
}
