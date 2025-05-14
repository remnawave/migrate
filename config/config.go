package config

import "github.com/alecthomas/kong"

type Config struct {
	PanelType         string `name:"panel-type" help:"Source panel type (e.g., marzban, marzneshin)" required:"true" default:"marzban" enum:"marzban,marzneshin" env:"PANEL_TYPE"`
	PanelURL          string `name:"panel-url" help:"Source panel URL" required:"true" env:"PANEL_URL"`
	PanelUsername     string `name:"panel-username" help:"Source panel admin username" required:"true" env:"PANEL_USERNAME"`
	PanelPassword     string `name:"panel-password" help:"Source panel admin password" required:"true" env:"PANEL_PASSWORD"`
	RemnawaveURL      string `name:"remnawave-url" help:"Destination panel URL" env:"REMNAWAVE_URL"`
	RemnawaveToken    string `name:"remnawave-token" help:"Destination panel API token" env:"REMNAWAVE_TOKEN"`
	BatchSize         int    `name:"batch-size" help:"Number of users to process in one batch" default:"100" env:"BATCH_SIZE"`
	LastUsers         int    `name:"last-users" help:"Only migrate last N users (0 means all users)" default:"0" env:"LAST_USERS"`
	PreferredStrategy string `name:"preferred-strategy" help:"Preferred traffic reset strategy for all users (NO_RESET, DAY, WEEK, MONTH). If set, overrides the user's original strategy" default:"" env:"PREFERRED_STRATEGY"`
	PreserveStatus    bool   `name:"preserve-status" help:"Preserve user status from source panel (if false, sets all users to ACTIVE)" default:"false" env:"PRESERVE_STATUS"`

	SourceHeadersRaw string `name:"source-headers" help:"Custom headers for source panel in key:value,key:value format" env:"SOURCE_HEADERS"`
	DestHeadersRaw   string `name:"dest-headers" help:"Custom headers for Remnawave panel in key:value,key:value format" env:"DEST_HEADERS"`

	// Эти поля будут заполняться вручную в main.go
	SourceHeaders map[string]string `kong:"-"`
	DestHeaders   map[string]string `kong:"-"`
}

func Parse(version string) *Config {
	var cfg Config
	kong.Parse(&cfg,
		kong.Name("remnawave-migrate"),
		kong.Description("Migrate users from various panels to Remnawave panel"),
		kong.Vars{"version": version},
	)
	return &cfg
}

