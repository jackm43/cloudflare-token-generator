package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"

	cftoken "github.com/jackmunro/cloudflare-token-generator"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		if err := runInit(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "generate":
		if err := runGenerate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "list-services":
		runListServices()
	case "list-zones":
		if err := runListZones(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: cloudflaretokengenerator <command> [args]

Commands:
  init                          Configure API token, account, and zone
  generate <service> <scope>    Generate a scoped API token
  list-services                 List available services
  list-zones                    List zones accessible by your token
  help                          Show this help

Scope:
  all                           All resources (all zones or configured account)
  <zone-id>                     Specific zone ID (zone-scoped services)
  <account-id>                  Specific account ID (account-scoped services)

Examples:
  cloudflaretokengenerator init
  cloudflaretokengenerator generate dns all
  cloudflaretokengenerator generate workers all
  cloudflaretokengenerator generate dns 023e105f4ecef8ad9ca31a8372d0c353`)
}

func readLine(r *bufio.Reader) string {
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

func runInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your Cloudflare API Token: ")
	apiToken := readLine(reader)
	if apiToken == "" {
		return fmt.Errorf("API token is required")
	}

	// Verify token
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}
	_, err = api.VerifyAPIToken(context.Background())
	if err != nil {
		return fmt.Errorf("token verification failed: %w", err)
	}
	fmt.Println("✓ Token verified")

	// Try to discover accounts
	var accountID string
	accounts, _, accErr := api.Accounts(context.Background(), cloudflare.AccountsListParams{})
	if accErr == nil && len(accounts) > 0 {
		fmt.Println("\nAvailable accounts:")
		for i, a := range accounts {
			fmt.Printf("  [%d] %s (%s)\n", i+1, a.Name, a.ID)
		}
		fmt.Print("\nSelect account (number) or enter Account ID: ")
		input := readLine(reader)
		if n, err := strconv.Atoi(input); err == nil && n >= 1 && n <= len(accounts) {
			accountID = accounts[n-1].ID
		} else {
			accountID = input
		}
	} else {
		fmt.Print("\nEnter your Account ID: ")
		accountID = readLine(reader)
	}
	if accountID == "" {
		return fmt.Errorf("account ID is required")
	}

	// Try to discover zones
	var zoneID string
	zones, zoneErr := api.ListZones(context.Background())
	if zoneErr == nil && len(zones) > 0 {
		fmt.Println("\nAvailable zones:")
		for i, z := range zones {
			fmt.Printf("  [%d] %s (%s)\n", i+1, z.Name, z.ID)
		}
		fmt.Print("\nSelect default zone (number), enter Zone ID, or press Enter to skip: ")
		input := readLine(reader)
		if input != "" {
			if n, err := strconv.Atoi(input); err == nil && n >= 1 && n <= len(zones) {
				zoneID = zones[n-1].ID
			} else {
				zoneID = input
			}
		}
	} else {
		fmt.Print("\nEnter default Zone ID (or press Enter to skip): ")
		zoneID = readLine(reader)
	}

	cfg := &cftoken.Config{
		APIToken:  apiToken,
		AccountID: accountID,
		ZoneID:    zoneID,
	}
	if err := cftoken.SaveConfig(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Println("\n✓ Config saved to ~/.goGenerateCFToken/config.yaml")
	return nil
}

func runGenerate() error {
	if len(os.Args) < 4 {
		return fmt.Errorf("usage: cloudflaretokengenerator generate <service> <scope>")
	}
	service := os.Args[2]
	scope := os.Args[3]

	cfg, err := cftoken.LoadConfig()
	if err != nil {
		return err
	}

	gen, err := cftoken.New(*cfg)
	if err != nil {
		return err
	}

	token, err := gen.Generate(service, scope)
	if err != nil {
		return err
	}

	fmt.Println(token)
	return nil
}

func runListServices() {
	fmt.Println("Available services:\n")
	fmt.Printf("  %-16s %-10s %s\n", "SERVICE", "SCOPE", "DESCRIPTION")
	fmt.Printf("  %-16s %-10s %s\n", "-------", "-----", "-----------")
	for _, svc := range cftoken.ListServices() {
		fmt.Printf("  %-16s %-10s %s\n", svc.Name, svc.ResourceScope, svc.Description)
	}
}

func runListZones() error {
	cfg, err := cftoken.LoadConfig()
	if err != nil {
		return err
	}

	gen, err := cftoken.New(*cfg)
	if err != nil {
		return err
	}

	zones, err := gen.DiscoverZones(context.Background())
	if err != nil {
		return fmt.Errorf("listing zones: %w", err)
	}

	if len(zones) == 0 {
		fmt.Println("No zones found (token may lack Zone Read permission)")
		return nil
	}

	fmt.Printf("%-40s %s\n", "ZONE ID", "NAME")
	fmt.Printf("%-40s %s\n", "-------", "----")
	for _, z := range zones {
		fmt.Printf("%-40s %s\n", z.ID, z.Name)
	}
	return nil
}
