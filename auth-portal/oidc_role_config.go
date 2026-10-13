package main

import (
	"auth-portal/configstore"
	"context"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const oidcRoleSection configstore.Section = "oidc-role-assignment"

type OIDCRolePolicy struct {
	ClientID          string        `json:"clientId"`
	ExportRoles       []string      `json:"exportRoles"`
	ExportPermissions bool          `json:"exportPermissions"`
	Plex              *PlexRoleRule `json:"plex,omitempty"`
}
type PlexRoleRule struct {
	OwnerUUID   string `json:"ownerUuid"`
	SectionKey  string `json:"sectionKey"`
	MatchRole   string `json:"matchRole"`
	DefaultRole string `json:"defaultRole"`
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
		if p.Plex == nil {
			continue
		}
		rule := p.Plex
		if rule.OwnerUUID == "" || rule.SectionKey == "" || rule.MatchRole == rule.DefaultRole {
			return errors.New("incomplete Plex assignment rule")
		}
		for _, name := range []string{rule.MatchRole, rule.DefaultRole} {
			d, ok := known[name]
			if !ok || isSystemRole(name) {
				return errors.New("automatic assignments require custom roles")
			}
			allowed := false
			for _, r := range p.ExportRoles {
				if r == name {
					allowed = true
				}
			}
			if !allowed {
				return errors.New("automatic role must be exported")
			}
			for _, perm := range d.Permissions {
				if isSystemPermission(perm) {
					return errors.New("automatic roles cannot grant system permissions")
				}
			}
		}
	}
	return nil
}

type plexRoleShares struct {
	XMLName xml.Name `xml:"MediaContainer"`
	Users   []struct {
		ID       string `xml:"userID,attr"`
		Sections []struct {
			Key    string `xml:"key,attr"`
			Shared string `xml:"shared,attr"`
		} `xml:"Section"`
	} `xml:"SharedServer"`
}

func plexShareMatches(body []byte, subject, owner, sectionKey string) (bool, error) {
	if sectionKey == "" || !strings.HasPrefix(subject, "plex-") {
		return false, errors.New("invalid Plex rule identity")
	}
	id := strings.TrimPrefix(subject, "plex-")
	n, err := strconv.ParseUint(id, 10, 64)
	if err != nil || n == 0 {
		return false, errors.New("invalid Plex identity")
	}
	var shares plexRoleShares
	if xml.Unmarshal(body, &shares) != nil {
		return false, errors.New("invalid Plex sharing response")
	}
	if subject == owner {
		return true, nil
	}
	count := 0
	matched := false
	for _, u := range shares.Users {
		if u.ID != id {
			continue
		}
		count++
		for _, s := range u.Sections {
			if s.Key == sectionKey && s.Shared == "1" {
				matched = true
			}
		}
	}
	if count != 1 {
		return false, errors.New("current Plex sharing required")
	}
	return matched, nil
}

var oidcRoleHTTPClient = &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

func fetchPlexRoleShares(rule PlexRoleRule) ([]byte, error) {
	if plexServerMachineID == "" || strings.ContainsAny(plexServerMachineID, "/?#\\") {
		return nil, errors.New("invalid Plex server")
	}
	owner, err := getUserByUUIDPreferred(rule.OwnerUUID)
	if err != nil || !owner.MediaToken.Valid || owner.MediaToken.String == "" {
		return nil, errors.New("Plex role-source credential unavailable")
	}
	req, err := http.NewRequest(http.MethodGet, "https://plex.tv/api/servers/"+plexServerMachineID+"/shared_servers", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Plex-Token", owner.MediaToken.String)
	req.Header.Set("Accept", "application/xml")
	resp, err := oidcRoleHTTPClient.Do(req)
	if err != nil {
		return nil, errors.New("Plex sharing unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("Plex sharing verification failed")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024+1))
	if err != nil || len(body) > 4*1024*1024 {
		return nil, errors.New("invalid Plex sharing response")
	}
	return body, nil
}
func assignOIDCPlexRole(userID int, clientID, role string, rule PlexRoleRule) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1),hashtext($2))`, string(oidcRoleSection), configstore.SectionDocumentKey); err != nil {
		return err
	}
	var raw []byte
	if err = tx.QueryRowContext(ctx, `SELECT value FROM app_config WHERE namespace=$1 AND key=$2`, string(oidcRoleSection), configstore.SectionDocumentKey).Scan(&raw); err != nil {
		return err
	}
	var current OIDCRoleConfig
	if json.Unmarshal(raw, &current) != nil {
		return errors.New("role configuration unavailable")
	}
	active := false
	for _, p := range current.Clients {
		if p.ClientID == clientID && p.Plex != nil && reflect.DeepEqual(*p.Plex, rule) && (role == rule.MatchRole || role == rule.DefaultRole) {
			active = true
		}
	}
	if !active {
		return errors.New("role policy changed; retry sign-in")
	}

	var roleID int
	if err = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name=$1`, role).Scan(&roleID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id=$1 AND source='oidc-plex' AND external_ref=$2 AND role_id<>$3`, userID, clientID, roleID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id,source,external_ref,granted_by) VALUES($1,$2,'oidc-plex',$3,'system:oidc-role-assignment') ON CONFLICT(user_id,role_id,source,external_ref) DO NOTHING`, userID, roleID, clientID); err != nil {
		return err
	}
	return tx.Commit()
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
	if policy.Plex != nil {
		body, err := fetchPlexRoleShares(*policy.Plex)
		if err != nil {
			return nil, err
		}
		matched, err := plexShareMatches(body, subjectForUser(user), policy.Plex.OwnerUUID, policy.Plex.SectionKey)
		if err != nil {
			return nil, err
		}
		role := policy.Plex.DefaultRole
		if matched {
			role = policy.Plex.MatchRole
		}
		if err = assignOIDCPlexRole(user.ID, clientID, role, *policy.Plex); err != nil {
			return nil, err
		}
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
	old, version, err := oidcRoleConfig()
	if err != nil {
		respondJSON(w, 500, map[string]any{"ok": false, "error": "role configuration unavailable"})
		return
	}
	if version != req.Version {
		respondJSON(w, 409, map[string]any{"ok": false, "error": "configuration changed; reload before saving"})
		return
	}
	if !saveOIDCSectionWithHook(w, r, oidcRoleSection, req.Config, req.Version, req.Reason, func(ctx context.Context, tx *sql.Tx) error { return pruneOIDCPlexRoles(ctx, tx, old, req.Config) }) {
		return
	}
	respondJSON(w, 200, map[string]any{"ok": true})
}

// Manual, LDAP and system roles remain global; generated role claims are client-scoped.
func roleNamesForClient(assignments []RoleAssignment, clientID string) []string {
	roles := []string{}
	for _, a := range assignments {
		if a.Source != "oidc-plex" || a.ExternalRef == clientID {
			roles = append(roles, a.Role)
		}
	}
	return normalizeDistinctStrings(roles)
}
func changedPlexPolicies(old, next OIDCRoleConfig) []string {
	removed := []string{}
	for _, previous := range old.Clients {
		if previous.Plex == nil {
			continue
		}
		var current *PlexRoleRule
		for _, p := range next.Clients {
			if p.ClientID == previous.ClientID {
				current = p.Plex
				break
			}
		}
		if !reflect.DeepEqual(previous.Plex, current) {
			removed = append(removed, previous.ClientID)
		}
	}
	return removed
}
func pruneOIDCPlexRoles(ctx context.Context, tx *sql.Tx, old, next OIDCRoleConfig) error {
	for _, id := range changedPlexPolicies(old, next) {
		if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE source='oidc-plex' AND external_ref=$1`, id); err != nil {
			return err
		}
	}
	return nil
}
