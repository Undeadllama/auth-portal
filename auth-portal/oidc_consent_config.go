package main

import (
	"auth-portal/configstore"
	"auth-portal/oauth"
	"net/http"
	"reflect"
)

const oidcConsentSection configstore.Section = "oidc-consent"

type TrustedOIDCClient struct {
	ClientID     string   `json:"clientId"`
	RedirectURIs []string `json:"redirectUris"`
	Scopes       []string `json:"scopes"`
}
type OIDCConsentConfig struct {
	Clients []TrustedOIDCClient `json:"clients"`
}

func oidcConsentConfig() (OIDCConsentConfig, int64, error) {
	cfg := OIDCConsentConfig{Clients: []TrustedOIDCClient{}}
	version, err := readOIDCSection(oidcConsentSection, &cfg)
	if cfg.Clients == nil {
		cfg.Clients = []TrustedOIDCClient{}
	}
	return cfg, version, err
}
func trustedConsentClient(cfg OIDCConsentConfig, client oauth.Client) bool {
	for _, p := range cfg.Clients {
		if p.ClientID == client.ClientID && len(p.RedirectURIs) > 0 && len(p.Scopes) > 0 &&
			reflect.DeepEqual(normalizeDistinctStrings(p.RedirectURIs), normalizeDistinctStrings(client.RedirectURIs)) &&
			reflect.DeepEqual(normalizeDistinctStrings(p.Scopes), normalizeDistinctStrings(client.Scopes)) {
			return true
		}
	}
	return false
}
func isTrustedOIDCClient(client oauth.Client) bool {
	cfg, _, err := oidcConsentConfig()
	return err == nil && trustedConsentClient(cfg, client)
}
func validateConsentClients(ids []string) error { return validateOIDCClientIDs(ids) }
func adminOIDCConsentConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodGet {
		cfg, version, err := oidcConsentConfig()
		if err != nil {
			respondJSON(w, 500, map[string]any{"ok": false, "error": "consent configuration unavailable"})
			return
		}
		respondJSON(w, 200, map[string]any{"ok": true, "config": cfg, "version": version})
		return
	}
	var req struct {
		ClientIDs []string `json:"clientIds"`
		Version   int64    `json:"version"`
		Reason    string   `json:"reason"`
	}
	if !decodeOIDCConfigRequest(w, r, &req) {
		return
	}
	if err := validateConsentClients(req.ClientIDs); err != nil {
		respondJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	cfg := OIDCConsentConfig{Clients: []TrustedOIDCClient{}}
	for _, id := range req.ClientIDs {
		c, err := oauthService.Client(r.Context(), id)
		if err != nil {
			respondJSON(w, 400, map[string]any{"ok": false, "error": "unknown OAuth client"})
			return
		}
		cfg.Clients = append(cfg.Clients, TrustedOIDCClient{c.ClientID, c.RedirectURIs, c.Scopes})
	}
	if !saveOIDCSection(w, r, oidcConsentSection, cfg, req.Version, req.Reason) {
		return
	}
	respondJSON(w, 200, map[string]any{"ok": true})
}
