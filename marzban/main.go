package main

import (
	"log"
	"strings"

	"marzban-migration-tool/config"
	"marzban-migration-tool/marzban"
	"marzban-migration-tool/migrator"
	"marzban-migration-tool/remnawave"
)

var (
	version = "unknown"
)

func main() {
	cfg := config.Parse(version)

	if cfg.MarzbanPassword == "" {
		log.Fatal("Marzban password is required")
	}
	if cfg.RemnawaveToken == "" {
		log.Fatal("Remnawave token is required")
	}

	// Validate PreferredStrategy if provided
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

	marzbanPanel := marzban.NewPanel(cfg.MarzbanURL)
	if err := marzbanPanel.Login(cfg.MarzbanUsername, cfg.MarzbanPassword); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	remnaPanel := remnawave.NewPanel(cfg.RemnawaveURL, cfg.RemnawaveToken)

	m := migrator.New(marzbanPanel, remnaPanel, cfg.PreferredStrategy, cfg.PreserveStatus)
	if err := m.MigrateUsers(cfg.BatchSize, cfg.LastUsers); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
