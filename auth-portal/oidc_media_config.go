package main

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"auth-portal/configstore"
)

const oidcMediaSection configstore.Section = "oidc-media-authorization"

var errOIDCMediaDenied = errors.New("authorized media server access required")

// Client admission is independent of the scopes requested by the caller.
// ScopeAliases preserves existing dynamic scopes during migration; new clients
// need only be selected in the administrator interface.
type OIDCMediaConfig struct {
	ClientIDs    []string `json:"clientIds"`
	ScopeAliases []string `json:"scopeAliases"`
}

func oidcMediaConfig() (OIDCMediaConfig, int64, error) {
	cfg := OIDCMediaConfig{ClientIDs: []string{}, ScopeAliases: []string{}}
	version, err := readOIDCSection(oidcMediaSection, &cfg)
	if cfg.ClientIDs == nil {
		cfg.ClientIDs = []string{}
	}
	if cfg.ScopeAliases == nil {
		cfg.ScopeAliases = []string{}
	}
	if err == nil {
		err = validateOIDCMediaConfig(cfg)
	}
	return cfg, version, err
}

func validateOIDCMediaConfig(cfg OIDCMediaConfig) error {
	if err := validateOIDCClientIDs(cfg.ClientIDs); err != nil {
		return err
	}
	if len(cfg.ScopeAliases) > 100 {
		return errors.New("too many media scope aliases")
	}
	seen := make(map[string]bool)
	for _, scope := range cfg.ScopeAliases {
		if scope == "" || strings.ContainsAny(scope, " \t\r\n") || isStandardOIDCScope(scope) || isSystemPermission(scope) || scope != normalizeRBACName(scope) || seen[scope] {
			return errors.New("media aliases must be unique custom scope names")
		}
		seen[scope] = true
	}
	return nil
}

func requireMediaAuthorization(uuid, username string) error {
	provider := activeProvider()
	if provider == nil {
		return errOIDCMediaDenied
	}
	authorized, err := provider.IsAuthorized(uuid, username)
	if err != nil || !authorized {
		return errOIDCMediaDenied
	}
	return nil
}

func enforceOIDCMediaPolicy(clientID, uuid, username string) error {
	cfg, _, err := oidcMediaConfig()
	if err != nil {
		return errors.New("media authorization configuration unavailable")
	}
	if containsScope(cfg.ClientIDs, clientID) {
		return requireMediaAuthorization(uuid, username)
	}
	return nil
}

func enforceOIDCMediaPolicyContext(ctx context.Context, clientID string) error {
	return enforceOIDCMediaPolicy(clientID, strings.TrimSpace(uuidFrom(ctx)), strings.TrimSpace(usernameFrom(ctx)))
}

func adminOIDCMediaConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodGet {
		cfg, version, err := oidcMediaConfig()
		if err != nil {
			respondJSON(w, 500, map[string]any{"ok": false, "error": "media configuration unavailable"})
			return
		}
		respondJSON(w, 200, map[string]any{"ok": true, "config": cfg, "version": version})
		return
	}
	var req struct {
		Config  OIDCMediaConfig `json:"config"`
		Version int64           `json:"version"`
		Reason  string          `json:"reason"`
	}
	if !decodeOIDCConfigRequest(w, r, &req) {
		return
	}
	if err := validateOIDCMediaConfig(req.Config); err != nil {
		respondJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	for _, id := range req.Config.ClientIDs {
		if _, err := oauthService.Client(r.Context(), id); err != nil {
			respondJSON(w, 400, map[string]any{"ok": false, "error": "unknown OAuth client"})
			return
		}
	}
	current, _, currentErr := oidcMediaConfig()
	if currentErr != nil || !reflect.DeepEqual(normalizeDistinctStrings(current.ScopeAliases), normalizeDistinctStrings(req.Config.ScopeAliases)) {
		granted, _, err := userHasPermission(uuidFrom(r.Context()), usernameFrom(r.Context()), permissionRBACWrite)
		if err != nil {
			respondJSON(w, 500, map[string]any{"ok": false, "error": "permission verification failed"})
			return
		}
		if !granted {
			respondJSON(w, 403, map[string]any{"ok": false, "error": "Access Control write permission is required to change compatibility scope aliases"})
			return
		}
	}
	if !saveOIDCSection(w, r, oidcMediaSection, req.Config, req.Version, req.Reason) {
		return
	}
	respondJSON(w, 200, map[string]any{"ok": true})
}
