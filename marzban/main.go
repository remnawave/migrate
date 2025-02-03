package main

import (
	"log"

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

	marzbanPanel := marzban.NewPanel(cfg.MarzbanURL)
	if err := marzbanPanel.Login(cfg.MarzbanUsername, cfg.MarzbanPassword); err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	remnaPanel := remnawave.NewPanel(cfg.RemnawaveURL, cfg.RemnawaveToken)

	m := migrator.New(marzbanPanel, remnaPanel, cfg.CalendarStrategy, cfg.PreserveStatus)
	if err := m.MigrateUsers(cfg.BatchSize, cfg.LastUsers); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
