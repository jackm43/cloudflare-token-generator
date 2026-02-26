//go:generate go run ./internal/generate

package cftoken

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"gopkg.in/yaml.v3"
)

const configDir = ".goGenerateCFToken"

// Config holds the stored credentials and defaults.
type Config struct {
	APIToken  string `yaml:"api_token"`
	AccountID string `yaml:"account_id"`
	ZoneID    string `yaml:"zone_id,omitempty"`
}

// Generator creates scoped Cloudflare API tokens.
type Generator struct {
	api       *cloudflare.API
	apiToken  string
	accountID string
	zoneID    string
}

// LoadConfig reads the config from ~/.goGenerateCFToken/config.yaml.
func LoadConfig() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(home, configDir, "config.yaml"))
	if err != nil {
		return nil, fmt.Errorf("config not found, run init first: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes the config to ~/.goGenerateCFToken/config.yaml.
func SaveConfig(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.yaml"), data, 0600)
}

// New creates a Generator from the given config.
func New(cfg Config) (*Generator, error) {
	api, err := cloudflare.NewWithAPIToken(cfg.APIToken)
	if err != nil {
		return nil, fmt.Errorf("creating cloudflare client: %w", err)
	}
	return &Generator{
		api:       api,
		apiToken:  cfg.APIToken,
		accountID: cfg.AccountID,
		zoneID:    cfg.ZoneID,
	}, nil
}

// Service convenience methods — each delegates to Generate.

func (g *Generator) DNS(scope string) (string, error)          { return g.Generate("dns", scope) }
func (g *Generator) Workers(scope string) (string, error)      { return g.Generate("workers", scope) }
func (g *Generator) R2(scope string) (string, error)           { return g.Generate("r2", scope) }
func (g *Generator) Pages(scope string) (string, error)        { return g.Generate("pages", scope) }
func (g *Generator) KV(scope string) (string, error)           { return g.Generate("kv", scope) }
func (g *Generator) Cache(scope string) (string, error)        { return g.Generate("cache", scope) }
func (g *Generator) Firewall(scope string) (string, error)     { return g.Generate("firewall", scope) }
func (g *Generator) SSL(scope string) (string, error)          { return g.Generate("ssl", scope) }
func (g *Generator) WAF(scope string) (string, error)          { return g.Generate("waf", scope) }
func (g *Generator) Stream(scope string) (string, error)       { return g.Generate("stream", scope) }
func (g *Generator) AI(scope string) (string, error)           { return g.Generate("ai", scope) }
func (g *Generator) D1(scope string) (string, error)           { return g.Generate("d1", scope) }
func (g *Generator) Queues(scope string) (string, error)       { return g.Generate("queues", scope) }
func (g *Generator) Images(scope string) (string, error)       { return g.Generate("images", scope) }
func (g *Generator) Tunnels(scope string) (string, error)      { return g.Generate("tunnels", scope) }
func (g *Generator) Zone(scope string) (string, error)         { return g.Generate("zone", scope) }
func (g *Generator) LoadBalancer(scope string) (string, error) { return g.Generate("loadbalancer", scope) }
func (g *Generator) PageRules(scope string) (string, error)    { return g.Generate("pagerules", scope) }

// Generate creates a Cloudflare API token for the given service and scope.
// Scope is "all" for all resources, or a specific zone/account ID.
func (g *Generator) Generate(service, scope string) (string, error) {
	return g.GenerateMulti([]string{service}, scope, "edit")
}

// GenerateMulti creates a single Cloudflare API token covering multiple services.
// Services are looked up by name. Scope is "all" for all resources, or a specific ID.
// Level is "read" for read-only permissions or "edit" for read+write permissions.
func (g *Generator) GenerateMulti(services []string, scope, level string) (string, error) {
	level = strings.ToLower(level)
	if level != "read" && level != "edit" {
		return "", fmt.Errorf("invalid permission level %q, must be \"read\" or \"edit\"", level)
	}

	var svcs []Service
	for _, s := range services {
		svc, ok := Services[strings.ToLower(strings.TrimSpace(s))]
		if !ok {
			return "", fmt.Errorf("unknown service %q, use ListServices() to see available services", s)
		}
		if len(filterPermissions(svc.Permissions, level)) == 0 {
			return "", fmt.Errorf("service %q does not support %q level (available: %s)",
				svc.Name, level, strings.Join(ServiceLevels(svc), ", "))
		}
		svcs = append(svcs, svc)
	}

	// Group services by resource scope to create correct policies.
	var zoneSvcs, accountSvcs []Service
	for _, svc := range svcs {
		if svc.ResourceScope == ResourceScopeZone {
			zoneSvcs = append(zoneSvcs, svc)
		} else {
			accountSvcs = append(accountSvcs, svc)
		}
	}

	var policies []cloudflare.APITokenPolicies

	if len(zoneSvcs) > 0 {
		resources, err := g.buildResources(zoneSvcs[0], scope)
		if err != nil {
			return "", err
		}
		var permGroups []cloudflare.APITokenPermissionGroups
		for _, svc := range zoneSvcs {
			for _, p := range filterPermissions(svc.Permissions, level) {
				permGroups = append(permGroups, cloudflare.APITokenPermissionGroups{ID: p.ID})
			}
		}
		policies = append(policies, cloudflare.APITokenPolicies{
			Effect:           "allow",
			Resources:        resources,
			PermissionGroups: permGroups,
		})
	}

	if len(accountSvcs) > 0 {
		resources, err := g.buildResources(accountSvcs[0], scope)
		if err != nil {
			return "", err
		}
		var permGroups []cloudflare.APITokenPermissionGroups
		for _, svc := range accountSvcs {
			for _, p := range filterPermissions(svc.Permissions, level) {
				permGroups = append(permGroups, cloudflare.APITokenPermissionGroups{ID: p.ID})
			}
		}
		policies = append(policies, cloudflare.APITokenPolicies{
			Effect:           "allow",
			Resources:        resources,
			PermissionGroups: permGroups,
		})
	}

	var names []string
	for _, svc := range svcs {
		names = append(names, svc.Name)
	}
	tokenName := fmt.Sprintf("%s-%s-%s", strings.Join(names, "-"), scope, level)

	return g.createToken(tokenName, policies)
}

func (g *Generator) createToken(name string, policies []cloudflare.APITokenPolicies) (string, error) {
	token := cloudflare.APIToken{
		Name:     name,
		Policies: policies,
	}

	result, err := g.api.CreateAPIToken(context.Background(), token)
	if err != nil {
		return "", fmt.Errorf("creating token: %w", err)
	}

	return result.Value, nil
}

// ServiceLevels returns the permission levels available for a service.
// Every service supports "edit" (all permissions). A service supports
// "read" only if it has at least one permission whose name contains "Read".
func ServiceLevels(svc Service) []string {
	for _, p := range svc.Permissions {
		if strings.Contains(strings.ToLower(p.Name), "read") {
			return []string{"read", "edit"}
		}
	}
	return []string{"edit"}
}

func filterPermissions(perms []Permission, level string) []Permission {
	if level == "edit" {
		return perms
	}
	var filtered []Permission
	for _, p := range perms {
		if strings.Contains(strings.ToLower(p.Name), "read") {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func (g *Generator) buildResources(svc Service, scope string) (map[string]interface{}, error) {
	resources := make(map[string]interface{})

	switch strings.ToLower(scope) {
	case "all":
		if svc.ResourceScope == ResourceScopeZone {
			resources["com.cloudflare.api.account.zone.*"] = "*"
		} else {
			if g.accountID == "" {
				return nil, fmt.Errorf("account_id required for account-scoped service %q with scope \"all\"", svc.Name)
			}
			resources["com.cloudflare.api.account."+g.accountID] = "*"
		}
	default:
		// Scope is a specific resource ID
		if svc.ResourceScope == ResourceScopeZone {
			resources["com.cloudflare.api.account.zone."+scope] = "*"
		} else {
			resources["com.cloudflare.api.account."+scope] = "*"
		}
	}

	return resources, nil
}

// permissionGroup represents a permission group returned by the Cloudflare API.
type permissionGroup struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

// fetchPermissionGroups fetches all available permission groups from the Cloudflare API.
func (g *Generator) fetchPermissionGroups() ([]permissionGroup, error) {
	req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/user/tokens/permission_groups", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+g.apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching permission groups: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Result  []permissionGroup `json:"result"`
		Success bool              `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding permission groups: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("API returned success=false (HTTP %d)", resp.StatusCode)
	}
	return result.Result, nil
}

// GodMode generates a single token with edit-level access to every service.
// It dynamically fetches all available permission groups from the Cloudflare API
// to ensure complete coverage.
func (g *Generator) GodMode() (string, error) {
	if g.accountID == "" {
		return "", fmt.Errorf("account_id required for godmode")
	}

	perms, err := g.fetchPermissionGroups()
	if err != nil {
		return "", err
	}

	var zonePerms, accountPerms []cloudflare.APITokenPermissionGroups
	for _, p := range perms {
		// Sub-tokens cannot manage other tokens.
		nameLower := strings.ToLower(p.Name)
		if strings.Contains(nameLower, "api token") {
			continue
		}
		scope := deriveScope(p.Scopes)
		switch scope {
		case "zone":
			zonePerms = append(zonePerms, cloudflare.APITokenPermissionGroups{ID: p.ID})
		case "account":
			accountPerms = append(accountPerms, cloudflare.APITokenPermissionGroups{ID: p.ID})
		}
	}

	var policies []cloudflare.APITokenPolicies

	if len(zonePerms) > 0 {
		policies = append(policies, cloudflare.APITokenPolicies{
			Effect:           "allow",
			Resources:        map[string]interface{}{"com.cloudflare.api.account.zone.*": "*"},
			PermissionGroups: zonePerms,
		})
	}

	if len(accountPerms) > 0 {
		policies = append(policies, cloudflare.APITokenPolicies{
			Effect:           "allow",
			Resources:        map[string]interface{}{"com.cloudflare.api.account." + g.accountID: "*"},
			PermissionGroups: accountPerms,
		})
	}

	return g.createToken("godmode", policies)
}

func deriveScope(scopes []string) string {
	for _, s := range scopes {
		if strings.Contains(s, "zone") {
			return "zone"
		}
		if strings.Contains(s, "account") {
			return "account"
		}
	}
	return ""
}

// ListServices returns all available service names sorted alphabetically.
func ListServices() []Service {
	var result []Service
	for _, svc := range Services {
		result = append(result, svc)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// DiscoverAccounts lists accounts accessible by the configured token.
func (g *Generator) DiscoverAccounts(ctx context.Context) ([]cloudflare.Account, error) {
	accounts, _, err := g.api.Accounts(ctx, cloudflare.AccountsListParams{})
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

// DiscoverZones lists zones accessible by the configured token.
func (g *Generator) DiscoverZones(ctx context.Context) ([]cloudflare.Zone, error) {
	zones, err := g.api.ListZones(ctx)
	if err != nil {
		return nil, err
	}
	return zones, nil
}
