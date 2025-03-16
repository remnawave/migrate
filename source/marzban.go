package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"remnawave-migrate/models"
)

type MarzbanPanel struct {
	client    *http.Client
	baseURL   string
	authToken string
}

func NewMarzbanPanel(baseURL string) *MarzbanPanel {
	return &MarzbanPanel{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (p *MarzbanPanel) Login(username, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	req, err := http.NewRequest("POST", p.baseURL+"/api/admin/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: status %d, body: %s", resp.StatusCode, body)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	p.authToken = tokenResp.AccessToken
	return nil
}

func (p *MarzbanPanel) GetUsers(offset, limit int) (*models.UsersResponse, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/api/users?offset=%d&limit=%d", p.baseURL, offset, limit),
		nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.authToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("getting users failed: status %d, body: %s",
			resp.StatusCode, body)
	}

	var marzbanResp models.MarzbanUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&marzbanResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	users := &models.UsersResponse{
		Users: make([]models.User, len(marzbanResp.Users)),
		Total: marzbanResp.Total,
	}

	for i, user := range marzbanResp.Users {
		users.Users[i] = models.User{
			MarzbanUser: user,
		}
	}

	return users, nil
}
