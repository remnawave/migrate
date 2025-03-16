package source

import (
	"fmt"

	"remnawave-migrate/models"
)

type SourcePanel interface {
	Login(username, password string) error

	GetUsers(offset, limit int) (*models.UsersResponse, error)
}

func Factory(panelType, baseURL string) (SourcePanel, error) {
	switch panelType {
	case "marzban":
		return NewMarzbanPanel(baseURL), nil
	case "marzneshin":
		return NewMarzneshinPanel(baseURL), nil
	default:
		return nil, fmt.Errorf("unsupported panel type: %s", panelType)
	}
}
