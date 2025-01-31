package remnawave

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"marzban-migration-tool/models"
)

type Panel struct {
	client  *http.Client
	baseURL string
	token   string
}

func NewPanel(baseURL, token string) *Panel {
	return &Panel{
		client:  &http.Client{},
		baseURL: baseURL,
		token:   token,
	}
}

func (p *Panel) CreateUser(req models.CreateUserRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.baseURL+"/api/users",
		strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		var apiErr ApiError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return fmt.Errorf("failed to parse error response: %w", err)
		}

		if apiErr.ErrorCode == "A019" {
			return &UserExistsError{
				Username: req.Username,
				ApiError: apiErr,
			}
		}
		return fmt.Errorf("creating user failed: status %d, body: %s",
			resp.StatusCode, body)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("creating user failed: status %d, body: %s",
			resp.StatusCode, body)
	}

	return nil
}
