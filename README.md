# Cloudflare Token Generator

Generate scoped Cloudflare API tokens from a bootstrap token. Works as both a Go SDK and CLI tool.

## Setup

```bash
go install github.com/jackm43/cloudflare-token-generator/cmd/cloudflaretokengenerator@latest
```

### Initialize config

```bash
cloudflaretokengenerator init
```

This prompts for your API token and account/zone details, saving to `~/.goGenerateCFToken/config.yaml`. If your token has Zone/Account Read permissions, available resources are auto-discovered.

## CLI Usage

```bash
# Generate a DNS token for all zones
cloudflaretokengenerator generate dns all

# Generate a Workers token for your account
cloudflaretokengenerator generate workers all

# Generate a DNS token for a specific zone
cloudflaretokengenerator generate dns <zone-id>

# List available services
cloudflaretokengenerator list-services

# List zones your token can see
cloudflaretokengenerator list-zones
```

## SDK Usage

```go
import cftoken "github.com/jackm43/cloudflare-token-generator"

// Load from ~/.goGenerateCFToken/config.yaml
cfg, _ := cftoken.LoadConfig()
gen, _ := cftoken.New(*cfg)

// Or configure directly
gen, _ := cftoken.New(cftoken.Config{
    APIToken:  "your-bootstrap-token",
    AccountID: "your-account-id",
})

// Generate tokens via service methods
dnsToken, _ := gen.DNS("all")
workersToken, _ := gen.Workers("all")
r2Token, _ := gen.R2("all")

// Or use the generic method
token, _ := gen.Generate("dns", "all")

// Scope to a specific zone
token, _ := gen.DNS("zone-id-here")
```

## Available Services

| Service | Scope | Description |
|---------|-------|-------------|
| `dns` | zone | DNS records management |
| `zone` | zone | Zone settings management |
| `cache` | zone | Cache purge |
| `firewall` | zone | Firewall services |
| `ssl` | zone | SSL and certificates |
| `waf` | zone | Zone WAF management |
| `loadbalancer` | zone | Load balancer management |
| `pagerules` | zone | Page rules management |
| `workers` | account | Workers scripts |
| `kv` | account | Workers KV storage |
| `r2` | account | Workers R2 storage |
| `pages` | account | Cloudflare Pages |
| `d1` | account | D1 database |
| `queues` | account | Cloudflare Queues |
| `ai` | account | Workers AI |
| `stream` | account | Cloudflare Stream |
| `images` | account | Cloudflare Images |
| `tunnels` | account | Cloudflare Tunnels |

## Bootstrap Token Requirements

Your bootstrap API token needs the **API Tokens Write** permission. For auto-discovery during `init`, it also needs **Account Read** and/or **Zone Read**.
