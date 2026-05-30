package source

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"remnawave-migrate/models"
)

type ThreeXUIPanel struct {
	client        *http.Client
	baseURL       string
	headers       map[string]string
	allUsersCache []models.User
	cacheLoaded   bool
}

type threeXUIClient struct {
	Email      string `json:"email"`
	ID         string `json:"id"`
	Password   string `json:"password"`
	Enable     bool   `json:"enable"`
	ExpiryTime int64  `json:"expiryTime"`
	CreatedAt  int64  `json:"created_at"`
	Comment    string `json:"comment"`
	LimitIp    int    `json:"limitIp"`
	TotalBytes int64  `json:"totalGB"` // Despite field called GB, the value is actually byte count
	Reset      int    `json:"reset"`
}

type threeXUIClientStatsItem struct {
	ID         int    `json:"id"`
	InboundID  int    `json:"inboundId"`
	Enable     bool   `json:"enable"`
	Email      string `json:"email"`
	UUID       string `json:"uuid"`
	SubID      string `json:"subId"`
	Up         int64  `json:"up"`
	Down       int64  `json:"down"`
	AllTime    int64  `json:"allTime"`
	ExpiryTime int64  `json:"expiryTime"`
	Total      int64  `json:"total"`
	Reset      int    `json:"reset"`
	LastOnline int64  `json:"lastOnline"`
}

// Combines two elements of the user ("client" in terms of 3x-ui) description
type user struct {
	client          threeXUIClient
	clientStatsItem threeXUIClientStatsItem
}

type inbound struct {
	ID           int                       `json:"id"`
	Protocol     string                    `json:"protocol"`
	Tag          string                    `json:"tag"`
	TrafficReset string                    `json:"trafficReset"`
	Settings     json.RawMessage           `json:"settings"`
	ClientStats  []threeXUIClientStatsItem `json:"clientStats"`
}

type apiResponse struct {
	Success bool      `json:"success"`
	Msg     string    `json:"msg"`
	Obj     []inbound `json:"obj"`
}

type threeXUIInboundSettings struct {
	Clients    []threeXUIClient `json:"clients"`
	Decryption string           `json:"decryption"`
	Encryption string           `json:"encryption"`
}

func NewThreeXUIPanel(baseURL string, headers map[string]string) *ThreeXUIPanel {
	jar, _ := cookiejar.New(nil)

	return &ThreeXUIPanel{
		client:  &http.Client{Jar: jar},
		baseURL: baseURL,
		headers: headers,
	}
}

func (p *ThreeXUIPanel) fetchCSRFToken() (string, bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/csrf-token", p.baseURL), nil)
	if err != nil {
		return "", false, fmt.Errorf("creating csrf token request: %w", err)
	}

	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("fetching csrf token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", false, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", false, fmt.Errorf("fetching csrf token failed: status %d, body: %s",
			resp.StatusCode, body)
	}

	var csrfResp struct {
		Success bool   `json:"success"`
		Obj     string `json:"obj"`
		Msg     string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&csrfResp); err != nil {
		return "", false, fmt.Errorf("decoding csrf token response: %w", err)
	}
	if !csrfResp.Success {
		return "", false, fmt.Errorf("csrf token request failed: %s", csrfResp.Msg)
	}
	if csrfResp.Obj == "" {
		return "", false, fmt.Errorf("csrf token response did not include token")
	}

	return csrfResp.Obj, true, nil
}

func (p *ThreeXUIPanel) Login(username, password string) error {
	csrfToken, csrfSupported, err := p.fetchCSRFToken()
	if err != nil {
		return err
	}

	loginData := url.Values{}
	loginData.Set("username", username)
	loginData.Set("password", password)
	loginData.Set("twoFactorCode", "")

	req, err := http.NewRequest("POST",
		fmt.Sprintf("%s/login", p.baseURL),
		strings.NewReader(loginData.Encode()))
	if err != nil {
		return fmt.Errorf("creating login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if csrfSupported {
		req.Header.Set("X-CSRF-Token", csrfToken)
	}
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: status %d, body: %s",
			resp.StatusCode, body)
	}

	return nil
}

func processUser(client threeXUIClient, clientStatsItem threeXUIClientStatsItem, inboundTrafficReset string) models.ProcessedUser {

	var expireTime time.Time
	if client.ExpiryTime > 0 {
		expireTime = time.Unix(client.ExpiryTime/1000, 0).UTC()
	} else {
		expireTime = time.Date(2099, 12, 31, 15, 13, 22, 214000000, time.UTC)
	}

	var status string
	if !client.Enable {
		status = "DISABLED"
	} else if expireTime.Before(time.Now().UTC()) {
		status = "EXPIRED"
	} else if clientStatsItem.LastOnline == 0 {
		status = "INACTIVE"
	} else {
		status = "ACTIVE"
	}

	var trafficResetStrategy string
	switch inboundTrafficReset {
	case "daily":
		trafficResetStrategy = "DAY"
	case "weekly":
		trafficResetStrategy = "WEEK"
	case "monthly":
		trafficResetStrategy = "MONTH"
	case "never":
	default:
		trafficResetStrategy = "NO_RESET"
	}

	var createdAtTime time.Time
	if client.CreatedAt > 0 {
		createdAtTime = time.Unix(client.CreatedAt/1000, 0).UTC()
	} else {
		createdAtTime = time.Now().UTC()
	}

	return models.ProcessedUser{
		Username:               client.Email,
		Status:                 status,
		VlessID:                "", //Protocol specific fields are set later
		TrojanPassword:         "",
		ShadowsocksPassword:    "",
		SubscriptionHash:       clientStatsItem.SubID,
		DataLimit:              client.TotalBytes,
		DataLimitResetStrategy: trafficResetStrategy,
		Note:                   client.Comment,
		Expire:                 expireTime.Format("2006-01-02T15:04:05.000Z"),
		CreatedAt:              createdAtTime.Format("2006-01-02T15:04:05.000Z"),
	}
}

func joinClientStats(clients []threeXUIClient, clientStats []threeXUIClientStatsItem) []user {
	clientsMap := make(map[string]threeXUIClient)
	for _, client := range clients {
		clientsMap[client.Email] = client
	}

	var users []user
	for _, stat := range clientStats {
		if client, ok := clientsMap[stat.Email]; ok {
			users = append(users, user{
				client:          client,
				clientStatsItem: stat,
			})
		}
	}
	return users
}

func getInboundUsers(inbound inbound) ([]models.ProcessedUser, error) {
	settingsObj, err := parseInboundSettings(inbound.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings for inbound %d: %w", inbound.ID, err)
	}

	users := joinClientStats(settingsObj.Clients, inbound.ClientStats)

	var processedUsers []models.ProcessedUser
	for _, user := range users {
		processedUser := processUser(user.client, user.clientStatsItem, inbound.TrafficReset)

		switch inbound.Protocol {
		case "shadowsocks":
			processedUser.ShadowsocksPassword = user.client.Password
		case "trojan":
			processedUser.TrojanPassword = user.client.Password
		case "vless":
		default:
			processedUser.VlessID = user.client.ID
		}

		processedUsers = append(processedUsers, processedUser)
	}
	return processedUsers, nil
}

func parseInboundSettings(raw json.RawMessage) (*threeXUIInboundSettings, error) {
	settingsJSON := strings.TrimSpace(string(raw))
	if settingsJSON == "" || settingsJSON == "null" {
		return &threeXUIInboundSettings{}, nil
	}

	if settingsJSON[0] == '"' {
		var settingsString string
		if err := json.Unmarshal(raw, &settingsString); err != nil {
			return nil, err
		}
		settingsJSON = strings.TrimSpace(settingsString)
		if settingsJSON == "" {
			return &threeXUIInboundSettings{}, nil
		}
	}

	var settingsObj threeXUIInboundSettings
	if err := json.Unmarshal([]byte(settingsJSON), &settingsObj); err != nil {
		return nil, err
	}

	return &settingsObj, nil
}

func fetchInbounds(p *ThreeXUIPanel) (*apiResponse, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/panel/api/inbounds/list", p.baseURL),
		nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

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
		return nil, fmt.Errorf("getting inbounds failed: status %d, body: %s",
			resp.StatusCode, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var apiResponse apiResponse

	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	if !apiResponse.Success {
		return nil, fmt.Errorf("API error: %s", apiResponse.Msg)
	}

	return &apiResponse, nil
}

func loadAllUsers(p *ThreeXUIPanel) ([]models.User, error) {
	if p.cacheLoaded {
		return p.allUsersCache, nil
	}

	apiResponse, err := fetchInbounds(p)
	if err != nil {
		return nil, err
	}

	var allUsers []models.ProcessedUser

	for _, inbound := range apiResponse.Obj {
		inboundUsers, err := getInboundUsers(inbound)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, inboundUsers...)
	}

	p.allUsersCache = make([]models.User, len(allUsers))
	for i, processedUser := range allUsers {
		p.allUsersCache[i] = models.User{
			ProcessedUser: processedUser,
		}
	}

	p.cacheLoaded = true
	return p.allUsersCache, nil
}

func (p *ThreeXUIPanel) GetUsers(offset, limit int) (*models.UsersResponse, error) {
	allUsers, err := loadAllUsers(p)
	if err != nil {
		return nil, err
	}

	total := len(allUsers)
	start := min(offset, total)
	end := min(offset+limit, total)

	if limit == 0 {
		return &models.UsersResponse{
			Users: allUsers,
			Total: total,
		}, nil
	}

	return &models.UsersResponse{
		Users: allUsers[start:end],
		Total: total,
	}, nil
}
