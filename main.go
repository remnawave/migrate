package main

import (
	"log"
	"strings"

	"remnawave-migrate/config"
	"remnawave-migrate/migrator"
	"remnawave-migrate/remnawave"
	"remnawave-migrate/source"
)

var (
	version = "unknown"
)

func main() {
	cfg := config.Parse(version)

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

	sourcePanel, err := source.Factory(cfg.PanelType, cfg.PanelURL)
	if err != nil {
		log.Fatalf("Failed to create source panel: %v", err)
	}

	if err := sourcePanel.Login(cfg.PanelUsername, cfg.PanelPassword); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	remnaPanel := remnawave.NewPanel(cfg.RemnawaveURL, cfg.RemnawaveToken)

	m := migrator.New(sourcePanel, remnaPanel, cfg.PreferredStrategy, cfg.PreserveStatus)
	if err := m.MigrateUsers(cfg.BatchSize, cfg.LastUsers); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
