---
name: generating-cftoken
description: "Generates scoped Cloudflare API tokens using the cloudflaretokengenerator CLI. Use when asked to create, generate, or manage Cloudflare API tokens, set up cftoken config, or list available Cloudflare services/zones."
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

> **Note:** `init` is interactive (reads from stdin). Do not run it via Bash tool. Instruct the user to run it manually.

### 2. Generate a Scoped Token

```bash
cloudflaretokengenerator generate <service> <scope>
```

- `<service>` — one of the services listed below
- `<scope>` — `all` (all resources) or a specific zone/account ID

The generated token is printed to stdout.

**Examples:**
```bash
# DNS token for all zones
cloudflaretokengenerator generate dns all

# Workers token for the configured account
cloudflaretokengenerator generate workers all

# DNS token for a specific zone
cloudflaretokengenerator generate dns 023e105f4ecef8ad9ca31a8372d0c353
```

### 3. List Available Services

```bash
cloudflaretokengenerator list-services
```

### 4. List Accessible Zones

```bash
cloudflaretokengenerator list-zones
```

## Available Services

### Zone-scoped
| Service | Description |
|---------|-------------|
| `dns` | DNS records management |
| `zone` | Zone settings management |
| `cache` | Cache purge |
| `firewall` | Firewall services |
| `ssl` | SSL and certificates |
| `waf` | Zone WAF management |
| `loadbalancer` | Load balancer management |
| `pagerules` | Page rules management |

### Account-scoped
| Service | Description |
|---------|-------------|
| `workers` | Workers scripts |
| `kv` | Workers KV storage |
| `r2` | Workers R2 storage |
| `pages` | Cloudflare Pages |
| `d1` | D1 database |
| `queues` | Cloudflare Queues |
| `ai` | Workers AI |
| `stream` | Cloudflare Stream |
| `images` | Cloudflare Images |
| `tunnels` | Cloudflare Tunnels |

## Key Details

- Config is stored at `~/.goGenerateCFToken/config.yaml` (YAML with `api_token`, `account_id`, `zone_id`)
- Zone-scoped services with `all` scope apply to all zones; account-scoped services with `all` require `account_id` in config
- Token names follow the pattern `<service>-<scope>` (e.g., `dns-all`, `workers-all`)
- The bootstrap token needs **API Tokens Write** permission
