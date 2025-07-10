package remnawave

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"remnawave-migrate/models"
)

type Panel struct {
	client  *http.Client
	baseURL string
	token   string
	headers map[string]string
}

func NewPanel(baseURL, token string, headers map[string]string) *Panel {
	return &Panel{
		client:  &http.Client{},
		baseURL: baseURL,
		token:   token,
		headers: headers,
	}
}

type Inbound struct {
	UUID string `json:"uuid"`
	Tag  string `json:"tag"`
}

type InboundsResponse struct {
	Response []Inbound `json:"response"`
}

func (p *Panel) GetInbounds() (map[string]string, error) {
	req, err := http.NewRequest("GET", p.baseURL+"/api/inbounds", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.token)
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get inbounds: status %d, body: %s", resp.StatusCode, body)
	}

	var inboundsResp InboundsResponse
	if err := json.NewDecoder(resp.Body).Decode(&inboundsResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	inbounds := make(map[string]string)
	for _, inbound := range inboundsResp.Response {
		inbounds[inbound.Tag] = inbound.UUID
	}

	return inbounds, nil
}

func (p *Panel) CreateUser(req models.CreateUserRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.baseURL+"/api/users", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+p.token)
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range p.headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusInternalServerError {
		body, _ := io.ReadAll(resp.Body)
		var apiErr ApiError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return fmt.Errorf("failed to parse error response: %w", err)
		}

		if apiErr.ErrorCode == "A019" || apiErr.ErrorCode == "A020" || apiErr.ErrorCode == "A021" || apiErr.ErrorCode == "A032" {
			return &UserExistsError{
				Username: req.Username,
				ApiError: apiErr,
			}
		}
		return fmt.Errorf("creating user failed: status %d, body: %s", resp.StatusCode, body)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("creating user failed: status %d, body: %s", resp.StatusCode, body)
	}

	return nil
}

