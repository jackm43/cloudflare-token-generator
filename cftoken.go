package cftoken

import (
	"context"
	"fmt"
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
	svc, ok := Services[strings.ToLower(service)]
	if !ok {
		return "", fmt.Errorf("unknown service %q, use ListServices() to see available services", service)
	}

	resources, err := g.buildResources(svc, scope)
	if err != nil {
		return "", err
	}

	var permGroups []cloudflare.APITokenPermissionGroups
	for _, p := range svc.Permissions {
		permGroups = append(permGroups, cloudflare.APITokenPermissionGroups{ID: p.ID})
	}

	token := cloudflare.APIToken{
		Name: fmt.Sprintf("%s-%s", svc.Name, scope),
		Policies: []cloudflare.APITokenPolicies{
			{
				Effect:           "allow",
				Resources:        resources,
				PermissionGroups: permGroups,
			},
		},
	}

	result, err := g.api.CreateAPIToken(context.Background(), token)
	if err != nil {
		return "", fmt.Errorf("creating token: %w", err)
	}

	return result.Value, nil
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
