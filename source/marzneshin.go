package source

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"remnawave-migrate/models"
)

type MarzneshinPanel struct {
	client    *http.Client
	baseURL   string
	authToken string
}

type MarzneshinUser struct {
	ID                     int     `json:"id"`
	Username               string  `json:"username"`
	ExpireStrategy         string  `json:"expire_strategy"`
	ExpireDate             *string `json:"expire_date"`
	DataLimit              *int64  `json:"data_limit"`
	DataLimitResetStrategy string  `json:"data_limit_reset_strategy"`
	Note                   *string `json:"note"`
	Key                    string  `json:"key"`
	Activated              bool    `json:"activated"`
	IsActive               bool    `json:"is_active"`
	Expired                bool    `json:"expired"`
	DataLimitReached       bool    `json:"data_limit_reached"`
	Enabled                bool    `json:"enabled"`
	SubscriptionURL        string  `json:"subscription_url"`
}

type MarzneshinUsersResponse struct {
	Items []MarzneshinUser `json:"items"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Size  int              `json:"size"`
	Pages int              `json:"pages"`
}

func NewMarzneshinPanel(baseURL string) *MarzneshinPanel {
	return &MarzneshinPanel{
		client:  &http.Client{},
		baseURL: baseURL,
	}
}

func (p *MarzneshinPanel) Login(username, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	req, err := http.NewRequest("POST", p.baseURL+"/api/admins/token",
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

func (p *MarzneshinPanel) GetUsers(offset, limit int) (*models.UsersResponse, error) {
	page := (offset / limit) + 1

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/api/users?page=%d&size=%d", p.baseURL, page, limit),
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

	var marzneshinResp MarzneshinUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&marzneshinResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	users := &models.UsersResponse{
		Users: make([]models.User, len(marzneshinResp.Items)),
		Total: marzneshinResp.Total,
	}

	for i, user := range marzneshinResp.Items {
		vlessID, trojanPassword, ssPassword, err := p.fetchUserProxies(user.Username, user.Key)
		if err != nil {
			return nil, fmt.Errorf("error fetching proxies for user %s: %w", user.Username, err)
		}

		processedUser := models.ProcessedUser{
			Username:               user.Username,
			VlessID:                vlessID,
			TrojanPassword:         trojanPassword,
			ShadowsocksPassword:    ssPassword,
			SubscriptionHash:       user.Key,
			DataLimitResetStrategy: strings.ToUpper(user.DataLimitResetStrategy),
			Note:                   getStringValue(user.Note),
		}

		if user.ExpireDate != nil {
			processedUser.Expire = *user.ExpireDate
		} else {
			farFuture := time.Date(2099, 12, 31, 15, 13, 22, 214000000, time.UTC).Format("2006-01-02T15:04:05.000Z")
			processedUser.Expire = farFuture
		}

		if user.DataLimit != nil {
			processedUser.DataLimit = *user.DataLimit
		}

		if !user.Enabled {
			processedUser.Status = "DISABLED"
		} else if user.Expired || user.DataLimitReached {
			processedUser.Status = "EXPIRED"
		} else if user.Activated && user.IsActive {
			processedUser.Status = "ACTIVE"
		} else {
			processedUser.Status = "INACTIVE"
		}

		users.Users[i] = models.User{
			ProcessedUser: processedUser,
		}
	}

	return users, nil
}

func (p *MarzneshinPanel) fetchUserProxies(username, key string) (string, string, string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/sub/%s/%s", p.baseURL, username, key), nil)
	if err != nil {
		return "", "", "", fmt.Errorf("creating subscription request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("fetching subscription: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", "", fmt.Errorf("subscription request failed: status %d, body: %s",
			resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", fmt.Errorf("reading subscription body: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return "", "", "", fmt.Errorf("decoding subscription content: %w", err)
	}

	configs := strings.Split(string(decoded), "\n")

	var vlessID, trojanPassword, ssPassword string

	for _, config := range configs {
		if strings.HasPrefix(config, "vless://") {
			vlessID = extractVlessID(config)
		} else if strings.HasPrefix(config, "trojan://") {
			trojanPassword = extractTrojanPassword(config)
		} else if strings.HasPrefix(config, "ss://") {
			ssPassword = extractShadowsocksPassword(config)
		}
	}

	return vlessID, trojanPassword, ssPassword, nil
}

func extractVlessID(config string) string {
	re := regexp.MustCompile(`vless://([^@]+)@`)
	matches := re.FindStringSubmatch(config)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractTrojanPassword(config string) string {
	re := regexp.MustCompile(`trojan://([^@]+)@`)
	matches := re.FindStringSubmatch(config)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractShadowsocksPassword(config string) string {
	re := regexp.MustCompile(`ss://([^#@]+)`)
	matches := re.FindStringSubmatch(config)
	if len(matches) > 1 {
		decoded, err := base64.StdEncoding.DecodeString(matches[1])
		if err != nil {
			return ""
		}

		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return ""
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
