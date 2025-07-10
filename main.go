package main

import (
	"log"
	"strings"

	"remnawave-migrate/config"
	"remnawave-migrate/migrator"
	"remnawave-migrate/remnawave"
	"remnawave-migrate/source"
	"remnawave-migrate/util"
)

var version = "unknown"

func main() {
	cfg := config.Parse(version)

	cfg.SourceHeaders = util.ParseHeaderMap(cfg.SourceHeadersRaw)
	cfg.DestHeaders = util.ParseHeaderMap(cfg.DestHeadersRaw)

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

	m := migrator.New(sourcePanel, remnaPanel, cfg.PreferredStrategy, cfg.PreserveInbounds, cfg.PreserveStatus, cfg.PreserveSubHash)
	if err := m.MigrateUsers(cfg.BatchSize, cfg.LastUsers); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
