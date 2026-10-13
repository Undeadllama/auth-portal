package main

import (
	"auth-portal/configstore"
	"errors"
	"net/http"
)

const oidcRoleSection configstore.Section = "oidc-role-assignment"

type OIDCRolePolicy struct {
	ClientID          string   `json:"clientId"`
	ExportRoles       []string `json:"exportRoles"`
	ExportPermissions bool     `json:"exportPermissions"`
}

type OIDCRoleConfig struct {
	Clients []OIDCRolePolicy `json:"clients"`
}

func oidcRoleConfig() (OIDCRoleConfig, int64, error) {
	cfg := OIDCRoleConfig{Clients: []OIDCRolePolicy{}}
	version, err := readOIDCSection(oidcRoleSection, &cfg)
	if cfg.Clients == nil {
		cfg.Clients = []OIDCRolePolicy{}
	}
	return cfg, version, err
}

func selectRoleClaims(assigned []string, policy OIDCRolePolicy, definitions []RoleDefinition) ([]string, []string) {
	roles := []string{}
	permissions := []string{}
	allowed := map[string]bool{}
	for _, r := range policy.ExportRoles {
		allowed[r] = true
	}
	for _, r := range assigned {
		if allowed[r] {
			roles = append(roles, r)
		}
	}
	if policy.ExportPermissions {
		for _, d := range definitions {
			for _, r := range roles {
				if d.Name == r {
					permissions = append(permissions, d.Permissions...)
				}
			}
		}
	}
	roles = normalizeDistinctStrings(roles)
	permissions = normalizePermissionList(permissions)
	if roles == nil {
		roles = []string{}
	}
	if permissions == nil {
		permissions = []string{}
	}
	return roles, permissions
}

func validateOIDCRoleConfig(cfg OIDCRoleConfig, definitions []RoleDefinition) error {
	known := map[string]RoleDefinition{}
	for _, r := range definitions {
		known[r.Name] = r
	}
	clients := map[string]bool{}
	if len(cfg.Clients) > 30 {
		return errors.New("too many client policies")
	}
	for _, p := range cfg.Clients {
		if p.ClientID == "" || clients[p.ClientID] {
			return errors.New("missing or duplicate client")
		}
		clients[p.ClientID] = true
		for _, r := range p.ExportRoles {
			if _, ok := known[r]; !ok {
				return errors.New("unknown exported role")
			}
		}

	}
	return nil
}

func configuredOIDCRoleClaims(clientID string, user User) (map[string]any, error) {
	cfg, _, err := oidcRoleConfig()
	if err != nil {
		return nil, err
	}
	var policy *OIDCRolePolicy
	for i := range cfg.Clients {
		if cfg.Clients[i].ClientID == clientID {
			policy = &cfg.Clients[i]
			break
		}
	}
	if policy == nil {
		return nil, nil
	}
	definitions, err := listRoleDefinitions()
	if err != nil {
		return nil, err
	}
	if err = validateOIDCRoleConfig(OIDCRoleConfig{Clients: []OIDCRolePolicy{*policy}}, definitions); err != nil {
		return nil, err
	}

	assignments, err := userRoleAssignments(user.MediaUUID.String, user.Username)
	if err != nil {
		return nil, err
	}
	assigned := roleNamesForClient(assignments, clientID)
	roles, permissions := selectRoleClaims(assigned, *policy, definitions)
	claims := map[string]any{"roles": roles}
	if policy.ExportPermissions {
		claims["permissions"] = permissions
	}
	return claims, nil
}

func adminOIDCRoleConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		cfg, version, err := oidcRoleConfig()
		if err != nil {
			respondJSON(w, 500, map[string]any{"ok": false, "error": "role configuration unavailable"})
			return
		}
		respondJSON(w, 200, map[string]any{"ok": true, "config": cfg, "version": version})
		return
	}
	var req struct {
		Config  OIDCRoleConfig `json:"config"`
		Version int64          `json:"version"`
		Reason  string         `json:"reason"`
	}
	if !decodeOIDCConfigRequest(w, r, &req) {
		return
	}
	defs, err := listRoleDefinitions()
	if err != nil {
		respondJSON(w, 500, map[string]any{"ok": false, "error": "roles unavailable"})
		return
	}
	if err = validateOIDCRoleConfig(req.Config, defs); err != nil {
		respondJSON(w, 400, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	for _, p := range req.Config.Clients {
		if _, err = oauthService.Client(r.Context(), p.ClientID); err != nil {
			respondJSON(w, 400, map[string]any{"ok": false, "error": "unknown OIDC client"})
			return
		}
	}
	if !saveOIDCSection(w, r, oidcRoleSection, req.Config, req.Version, req.Reason) {
		return
	}
	respondJSON(w, 200, map[string]any{"ok": true})
}

func roleNamesForClient(assignments []RoleAssignment, clientID string) []string {
	roles := []string{}
	for _, a := range assignments {
		if a.Source != "oidc-plex" || a.ExternalRef == clientID {
			roles = append(roles, a.Role)
		}
	}
	return normalizeDistinctStrings(roles)
}
