package models

import (
	"remnawave-migrate/util"
	"strings"
	"time"
)

type UsersResponse struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}

type User struct {
	MarzbanUser   MarzbanUser
	ProcessedUser ProcessedUser
}

func (u *User) Process() ProcessedUser {
	if u.ProcessedUser.Username != "" {
		return u.ProcessedUser
	}

	return u.MarzbanUser.Process()
}

type MarzbanProxies struct {
	Vless struct {
		ID   string `json:"id"`
		Flow string `json:"flow"`
	} `json:"vless"`
	Trojan struct {
		Password string `json:"password"`
		Flow     string `json:"flow"`
	} `json:"trojan"`
	Shadowsocks struct {
		Password string `json:"password"`
		Method   string `json:"method"`
	} `json:"shadowsocks"`
}

type MarzbanUser struct {
	Proxies                MarzbanProxies `json:"proxies"`
	Expire                 int64          `json:"expire"`
	DataLimit              int64          `json:"data_limit"`
	DataLimitResetStrategy string         `json:"data_limit_reset_strategy"`
	Note                   string         `json:"note"`
	Username               string         `json:"username"`
	Status                 string         `json:"status"`
	SubscriptionURL        string         `json:"subscription_url"`
}

type MarzbanUsersResponse struct {
	Users []MarzbanUser `json:"users"`
	Total int           `json:"total"`
}

type ProcessedUser struct {
	Expire                 string `json:"expire"`
	DataLimit              int64  `json:"data_limit"`
	DataLimitResetStrategy string `json:"data_limit_reset_strategy"`
	Note                   string `json:"note"`
	Username               string `json:"username"`
	Status                 string `json:"status"`
	VlessID                string `json:"vless_id"`
	TrojanPassword         string `json:"trojan_password"`
	ShadowsocksPassword    string `json:"shadowsocks_password"`
	SubscriptionHash       string `json:"subscription_hash"`
}

func (u *MarzbanUser) Process() ProcessedUser {
	var expireTime time.Time
	if u.Expire > 0 {
		expireTime = time.Unix(u.Expire, 0).UTC()
	} else {
		expireTime = time.Date(2099, 12, 31, 15, 13, 22, 214000000, time.UTC)
	}

	subscriptionHash := ""
	if u.SubscriptionURL != "" {
		parts := strings.Split(u.SubscriptionURL, "/")
		if len(parts) > 0 {
			subscriptionHash = parts[len(parts)-1]
		}
	}

	return ProcessedUser{
		Expire:                 expireTime.Format("2006-01-02T15:04:05.000Z"),
		DataLimit:              u.DataLimit,
		DataLimitResetStrategy: u.DataLimitResetStrategy,
		Note:                   u.Note,
		Username:               u.Username,
		Status:                 u.Status,
		VlessID:                u.Proxies.Vless.ID,
		TrojanPassword:         u.Proxies.Trojan.Password,
		ShadowsocksPassword:    u.Proxies.Shadowsocks.Password,
		SubscriptionHash:       subscriptionHash,
	}
}

type CreateUserRequest struct {
	Username             string  `json:"username"`
	Status               string  `json:"status"`
	ShortUUID            *string `json:"shortUuid,omitempty"`
	TrojanPassword       *string `json:"trojanPassword,omitempty"`
	VlessUUID            *string `json:"vlessUuid,omitempty"`
	SsPassword           *string `json:"ssPassword,omitempty"`
	TrafficLimitBytes    int64   `json:"trafficLimitBytes"`
	TrafficLimitStrategy string  `json:"trafficLimitStrategy"`
	ExpireAt             string  `json:"expireAt"`
	Description          string  `json:"description"`
	ActivateAllInbounds  bool    `json:"activateAllInbounds"`
}

func (p *ProcessedUser) ToCreateUserRequest(preferredStrategy string, preserveStatus bool) CreateUserRequest {
	strategy := strings.ToUpper(p.DataLimitResetStrategy)

	if strategy == "YEAR" {
		strategy = "NO_RESET"
	}

	if preferredStrategy != "" {
		strategy = preferredStrategy
	}

	status := "ACTIVE"
	if preserveStatus && strings.ToLower(p.Status) != "on_hold" {
		status = strings.ToUpper(p.Status)
	}

	validUsername := util.SanitizeUsername(p.Username)

	req := CreateUserRequest{
		Username:             validUsername,
		Status:               status,
		TrafficLimitBytes:    p.DataLimit,
		TrafficLimitStrategy: strategy,
		ExpireAt:             p.Expire,
		Description:          p.Note,
		ActivateAllInbounds:  true,
	}

	if p.SubscriptionHash != "" {
		req.ShortUUID = strPtr(p.SubscriptionHash)
	}
	if p.TrojanPassword != "" {
		req.TrojanPassword = strPtr(p.TrojanPassword)
	}
	if p.VlessID != "" {
		req.VlessUUID = strPtr(p.VlessID)
	}
	if p.ShadowsocksPassword != "" {
		req.SsPassword = strPtr(p.ShadowsocksPassword)
	}

	return req
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
