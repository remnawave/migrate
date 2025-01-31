package config

import "github.com/alecthomas/kong"

type Config struct {
	MarzbanURL       string `name:"marzban-url" help:"Marzban panel URL" required:"true" env:"MARZBAN_URL"`
	MarzbanUsername  string `name:"marzban-username" help:"Marzban admin username" required:"true" env:"MARZBAN_USERNAME"`
	MarzbanPassword  string `name:"marzban-password" help:"Marzban admin password" required:"true" env:"MARZBAN_PASSWORD"`
	RemnawaveURL     string `name:"remnawave-url" help:"Destination panel URL" env:"REMNAWAVE_URL"`
	RemnawaveToken   string `name:"remnawave-token" help:"Destination panel API token" env:"REMNAWAVE_TOKEN"`
	BatchSize        int    `name:"batch-size" help:"Number of users to process in one batch" default:"100" env:"BATCH_SIZE"`
	LastUsers        int    `name:"last-users" help:"Only migrate last N users (0 means all users)" default:"0" env:"LAST_USERS"`
	CalendarStrategy bool   `name:"calendar-strategy" help:"Force CALENDAR_MONTH reset strategy for all users" default:"false" env:"CALENDAR_STRATEGY"`
}

func Parse(version string) *Config {
	var cfg Config
	kong.Parse(&cfg,
		kong.Name("marzban-migration-tool"),
		kong.Description("Migrate users from Marzban panel to Remnawave panel"),
		kong.Vars{"version": version},
	)
	return &cfg
}
