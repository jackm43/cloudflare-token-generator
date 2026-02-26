---
name: generating-cftoken
description: Generates scoped Cloudflare API tokens using the cloudflaretokengenerator CLI. Use when asked to create, generate, or manage Cloudflare API tokens, set up cftoken config, or list available Cloudflare services/zones.
---

# Generating Cloudflare Tokens with the CLI

Generate scoped Cloudflare API tokens from a bootstrap token using the `cloudflaretokengenerator` CLI.

## Installation

```bash
go install github.com/jackm43/cloudflare-token-generator/cmd/cloudflaretokengenerator@latest
```

## Workflow

### 1. Initialize Configuration (First-time Setup)

Run the interactive init command to store credentials at `~/.goGenerateCFToken/config.yaml`:

```bash
cloudflaretokengenerator init
```

This prompts for:
- **API Token** — must have **API Tokens Write** permission; optionally **Account Read** and **Zone Read** for auto-discovery
- **Account ID** — selected from discovered accounts or entered manually
- **Zone ID** (optional) — default zone for zone-scoped services

`init` is interactive (reads from stdin). Do not run it via Bash tool. Instruct the user to run it manually.

### 2. Generate a Scoped Token

```bash
cloudflaretokengenerator generate <services> <scope> [level]
```

- `<services>` — comma-separated list of services (e.g. `workers,kv,d1`)
- `<scope>` — `all` (all resources) or a specific zone/account ID
- `[level]` — `edit` (read+write, default) or `read` (read-only)

The generated token is printed to stdout.

**Examples:**
```bash
# DNS token for all zones
cloudflaretokengenerator generate dns all

# Workers + KV token with edit permissions
cloudflaretokengenerator generate workers,kv all edit

# Workers + KV + D1 read-only token
cloudflaretokengenerator generate workers,kv,d1 all read

# DNS token for a specific zone
cloudflaretokengenerator generate dns 023e105f4ecef8ad9ca31a8372d0c353
```

### 3. God Mode — All Services Token

```bash
cloudflaretokengenerator godmode
```

Generates a single token with **edit** (read+write) access to **every** available service, scoped to `all` resources. No additional arguments needed.

### 4. List Available Services

```bash
cloudflaretokengenerator list-services
```

Shows each service's name, resource scope, supported permission levels, and description. Use this to check which levels (`read`, `edit`) a service supports before generating.

### 5. List Accessible Zones

```bash
cloudflaretokengenerator list-zones
```

## Available Services

### Zone-scoped
| Service | Levels | Description |
|---------|--------|-------------|
| `dns` | read, edit | DNS records management |
| `zone` | read, edit | Zone settings management |
| `cache` | edit | Cache purge |
| `firewall` | read, edit | Firewall services |
| `ssl` | read, edit | SSL and certificates |
| `waf` | read, edit | Zone WAF management |
| `loadbalancer` | read, edit | Load balancer management |
| `pagerules` | read, edit | Page rules management |

### Account-scoped
| Service | Levels | Description |
|---------|--------|-------------|
| `workers` | read, edit | Workers scripts |
| `kv` | read, edit | Workers KV storage |
| `r2` | read, edit | Workers R2 storage |
| `pages` | read, edit | Cloudflare Pages |
| `d1` | read, edit | D1 database |
| `queues` | read, edit | Cloudflare Queues |
| `ai` | read, edit | Workers AI |
| `stream` | read, edit | Cloudflare Stream |
| `images` | read, edit | Cloudflare Images |
| `tunnels` | read, edit | Cloudflare Tunnels |

## Key Details

- Config is stored at `~/.goGenerateCFToken/config.yaml` (YAML with `api_token`, `account_id`, `zone_id`)
- Multiple services can be combined in a single token (e.g. `workers,kv,d1`)
- Mixed-scope services (zone + account) are grouped into separate policies automatically
- Zone-scoped services with `all` scope apply to all zones; account-scoped services with `all` require `account_id` in config
- Token names follow the pattern `<services>-<scope>-<level>` (e.g., `workers-kv-all-edit`)
- Not all services support `read` level — use `list-services` to check; requesting an unsupported level returns an error
- The bootstrap token needs **API Tokens Write** permission
- Permission IDs in `services.go` are auto-generated from the Cloudflare API via `go generate` / the `update-services` GitHub Actions workflow
