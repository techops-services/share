package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/techops-services/share/internal/client"
	"github.com/techops-services/share/internal/config"
)

func newInitCmd(f *flags) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Set up share configuration",
		Long:  "Interactive setup wizard that creates ~/.config/share/config.toml.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(f)
		},
	}
}

func runInit(f *flags) error {
	reader := bufio.NewReader(os.Stdin)

	cfgPath := f.configPath
	if cfgPath == "" {
		cfgPath = config.DefaultConfigPath()
	}

	fmt.Println("share init")
	fmt.Println()

	// Endpoint.
	fmt.Printf("API endpoint [https://sh.techops.services]: ")
	endpoint, _ := reader.ReadString('\n')
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		endpoint = "https://sh.techops.services"
	}

	// API key.
	fmt.Printf("API key (optional, press Enter to skip): ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	// Default TTL.
	fmt.Printf("Default TTL [24h]: ")
	ttl, _ := reader.ReadString('\n')
	ttl = strings.TrimSpace(ttl)
	if ttl == "" {
		ttl = "24h"
	}

	// Validate TTL.
	if _, err := config.ParseTTL(ttl); err != nil {
		return fmt.Errorf("invalid TTL: %w", err)
	}

	// Verify connection.
	fmt.Printf("\nVerifying connection to %s... ", endpoint)
	c := client.New(client.WithEndpoint(endpoint))
	health, err := c.Health()
	if err != nil {
		fmt.Println("FAILED")
		return fmt.Errorf("could not connect to %s: %w", endpoint, err)
	}
	fmt.Printf("OK (v%s)\n", health.Version)

	// Save config.
	cfg := &config.Config{
		Endpoint:   endpoint,
		APIKey:     apiKey,
		DefaultTTL: ttl,
		Clipboard:  true,
	}

	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("\nConfig saved to %s\n", cfgPath)
	return nil
}
