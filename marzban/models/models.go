package models

import (
	"strings"
	"time"
)

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

func (p *ProcessedUser) ToCreateUserRequest(forceMonthlyReset bool, preserveStatus bool) CreateUserRequest {
	strategy := strings.ToUpper(p.DataLimitResetStrategy)
	if forceMonthlyReset {
		strategy = "CALENDAR_MONTH"
	}

	status := "ACTIVE"
	if preserveStatus {
		status = strings.ToUpper(p.Status)
	}

	validUsername := ensureValidUsername(p.Username)

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

type MarzbanUsersResponse struct {
	Users []MarzbanUser `json:"users"`
	Total int           `json:"total"`
}

func ensureValidUsername(username string) string {
	if len(username) < 6 {
		return username + strings.Repeat("_", 6-len(username))
	}
	return username
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
