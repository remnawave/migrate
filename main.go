package main

import (
	"log"
	"os"
	"strings"

	"remnawave-migrate/config"
	"remnawave-migrate/migrator"
	"remnawave-migrate/remnawave"
	"remnawave-migrate/source"
)

var version = "unknown"

func main() {
	cfg := config.Parse(version)

	cfg.SourceHeaders = parseHeaderMap(cfg.SourceHeadersRaw)
	cfg.DestHeaders = parseHeaderMap(cfg.DestHeadersRaw)

	if cfg.RemnawaveToken != "" {
		if _, exists := cfg.DestHeaders["Authorization"]; !exists {
			cfg.DestHeaders["Authorization"] = "Bearer " + cfg.RemnawaveToken
		}
	}

	if cfg.PanelPassword == "" {
		log.Fatal("Panel password is required")
	}
	if cfg.RemnawaveToken == "" {
		log.Fatal("Remnawave token is required")
	}

	if cfg.PreferredStrategy != "" {
		validStrategies := map[string]bool{
			"NO_RESET": true,
			"DAY":      true,
			"WEEK":     true,
			"MONTH":    true,
		}

		strategy := strings.ToUpper(cfg.PreferredStrategy)
		if !validStrategies[strategy] {
			log.Fatalf("Invalid preferred-strategy value: %s. Must be one of: NO_RESET, DAY, WEEK, MONTH", cfg.PreferredStrategy)
		}
		cfg.PreferredStrategy = strategy
	}

	log.Printf("Starting migration from %s panel...", cfg.PanelType)

	sourcePanel, err := source.Factory(cfg.PanelType, cfg.PanelURL, cfg.SourceHeaders)
	if err != nil {
		log.Fatalf("Failed to create source panel: %v", err)
	}

	if err := sourcePanel.Login(cfg.PanelUsername, cfg.PanelPassword); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	remnaPanel := remnawave.NewPanel(cfg.RemnawaveURL, cfg.RemnawaveToken, cfg.DestHeaders)

	m := migrator.New(sourcePanel, remnaPanel, cfg.PreferredStrategy, cfg.PreserveStatus)
	if err := m.MigrateUsers(cfg.BatchSize, cfg.LastUsers); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}

func parseHeaderMap(raw string) map[string]string {
	headers := make(map[string]string)
	if raw == "" {
		return headers
	}
	for _, pair := range strings.Split(raw, ",") {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return headers
}

func getEnvOrArg(envKey, flagKey string) string {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, flagKey+"=") {
			return strings.SplitN(arg, "=", 2)[1]
		}
	}
	return os.Getenv(envKey)
}

