package source

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestThreeXUILoginUsesCSRFSessionCookieAndKeepsLoginSession(t *testing.T) {
	const (
		csrfToken     = "csrf-token-123"
		anonymousSID  = "anonymous-session"
		authorizedSID = "authorized-session"
	)

	var sawCSRFTokenRequest bool
	var sawLoginRequest bool
	var sawInboundsRequest bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/csrf-token":
			sawCSRFTokenRequest = true
			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: anonymousSID, Path: "/"})
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": true,
				"obj":     csrfToken,
			})

		case "/login":
			sawLoginRequest = true

			if r.Method != http.MethodPost {
				t.Errorf("login method = %s, want POST", r.Method)
				http.Error(w, "bad method", http.StatusMethodNotAllowed)
				return
			}
			if got := r.Header.Get("X-CSRF-Token"); got != csrfToken {
				http.Error(w, "missing csrf token", http.StatusForbidden)
				return
			}
			cookie, err := r.Cookie("3x-ui")
			if err != nil || cookie.Value != anonymousSID {
				http.Error(w, "missing anonymous session cookie", http.StatusForbidden)
				return
			}
			if err := r.ParseForm(); err != nil {
				t.Errorf("ParseForm returned error: %v", err)
				http.Error(w, "bad form", http.StatusBadRequest)
				return
			}
			if r.Form.Get("username") != "admin" || r.Form.Get("password") != "secret" {
				http.Error(w, "bad credentials", http.StatusUnauthorized)
				return
			}

			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: authorizedSID, Path: "/"})
			w.WriteHeader(http.StatusOK)

		case "/panel/api/inbounds/list":
			sawInboundsRequest = true

			cookie, err := r.Cookie("3x-ui")
			if err != nil || cookie.Value != authorizedSID {
				http.Error(w, "missing authorized session cookie", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"obj":[]}`))

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	panel := NewThreeXUIPanel(server.URL, map[string]string{"X-Test-Header": "present"})

	if err := panel.Login("admin", "secret"); err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if _, err := panel.GetUsers(0, 100); err != nil {
		t.Fatalf("GetUsers returned error: %v", err)
	}

	if !sawCSRFTokenRequest {
		t.Fatal("expected /csrf-token to be requested before login")
	}
	if !sawLoginRequest {
		t.Fatal("expected /login to be requested")
	}
	if !sawInboundsRequest {
		t.Fatal("expected /panel/api/inbounds/list to be requested")
	}
}

func TestThreeXUILoginReportsCSRFTokenFetchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/csrf-token" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		http.Error(w, "csrf disabled", http.StatusInternalServerError)
	}))
	defer server.Close()

	panel := NewThreeXUIPanel(server.URL, nil)

	err := panel.Login("admin", "secret")
	if err == nil {
		t.Fatal("expected login to fail when csrf token fetch fails")
	}
	if !strings.Contains(err.Error(), "csrf token") {
		t.Fatalf("error %q does not mention csrf token", err)
	}
}

func TestThreeXUILoginFallsBackForLegacyPanelWithoutCSRFTokenEndpoint(t *testing.T) {
	const authorizedSID = "legacy-authorized-session"

	var sawCSRFTokenRequest bool
	var sawLoginRequest bool
	var sawInboundsRequest bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/csrf-token":
			sawCSRFTokenRequest = true
			http.NotFound(w, r)

		case "/login":
			sawLoginRequest = true

			if got := r.Header.Get("X-CSRF-Token"); got != "" {
				http.Error(w, "legacy login did not expect csrf token", http.StatusBadRequest)
				return
			}
			if err := r.ParseForm(); err != nil {
				t.Errorf("ParseForm returned error: %v", err)
				http.Error(w, "bad form", http.StatusBadRequest)
				return
			}
			if r.Form.Get("username") != "admin" || r.Form.Get("password") != "secret" {
				http.Error(w, "bad credentials", http.StatusUnauthorized)
				return
			}

			http.SetCookie(w, &http.Cookie{Name: "3x-ui", Value: authorizedSID, Path: "/"})
			w.WriteHeader(http.StatusOK)

		case "/panel/api/inbounds/list":
			sawInboundsRequest = true

			cookie, err := r.Cookie("3x-ui")
			if err != nil || cookie.Value != authorizedSID {
				http.Error(w, "missing legacy login session cookie", http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"obj":[]}`))

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	panel := NewThreeXUIPanel(server.URL, nil)

	if err := panel.Login("admin", "secret"); err != nil {
		t.Fatalf("Login returned error: %v", err)
	}

	if _, err := panel.GetUsers(0, 100); err != nil {
		t.Fatalf("GetUsers returned error: %v", err)
	}

	if !sawCSRFTokenRequest {
		t.Fatal("expected /csrf-token to be attempted before legacy fallback")
	}
	if !sawLoginRequest {
		t.Fatal("expected /login to be requested")
	}
	if !sawInboundsRequest {
		t.Fatal("expected /panel/api/inbounds/list to be requested")
	}
}

func TestThreeXUIAPIResponseAcceptsObjectSettings(t *testing.T) {
	body := []byte(`{
		"success": true,
		"obj": [
			{
				"id": 1,
				"protocol": "trojan",
				"trafficReset": "monthly",
				"settings": {
					"clients": [
						{
							"email": "alice",
							"password": "trojan-pass",
							"enable": true,
							"expiryTime": 4102444800000,
							"created_at": 1700000000000,
							"comment": "migrated user",
							"totalGB": 1234
						}
					]
				},
				"clientStats": [
					{
						"email": "alice",
						"subId": "alice-subhash",
						"lastOnline": 1700000000000
					}
				]
			}
		]
	}`)

	var response apiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	users, err := getInboundUsers(response.Obj[0])
	if err != nil {
		t.Fatalf("getInboundUsers returned error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("got %d users, want 1", len(users))
	}

	user := users[0]
	if user.Username != "alice" {
		t.Fatalf("Username = %q, want alice", user.Username)
	}
	if user.TrojanPassword != "trojan-pass" {
		t.Fatalf("TrojanPassword = %q, want trojan-pass", user.TrojanPassword)
	}
	if user.DataLimit != 1234 {
		t.Fatalf("DataLimit = %d, want 1234", user.DataLimit)
	}
	if user.DataLimitResetStrategy != "MONTH" {
		t.Fatalf("DataLimitResetStrategy = %q, want MONTH", user.DataLimitResetStrategy)
	}
	if user.Status != "ACTIVE" {
		t.Fatalf("Status = %q, want ACTIVE", user.Status)
	}
	if user.SubscriptionHash != "alice-subhash" {
		t.Fatalf("SubscriptionHash = %q, want alice-subhash", user.SubscriptionHash)
	}
}

func TestThreeXUIAPIResponseAcceptsStringSettings(t *testing.T) {
	body := []byte(`{
		"success": true,
		"obj": [
			{
				"id": 1,
				"protocol": "shadowsocks",
				"trafficReset": "weekly",
				"settings": "{\"clients\":[{\"email\":\"bob\",\"password\":\"ss-pass\",\"enable\":true,\"expiryTime\":4102444800000,\"created_at\":1700000000000,\"comment\":\"legacy user\",\"totalGB\":5678}]}",
				"clientStats": [
					{
						"email": "bob",
						"lastOnline": 1700000000000
					}
				]
			}
		]
	}`)

	var response apiResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	users, err := getInboundUsers(response.Obj[0])
	if err != nil {
		t.Fatalf("getInboundUsers returned error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("got %d users, want 1", len(users))
	}

	user := users[0]
	if user.Username != "bob" {
		t.Fatalf("Username = %q, want bob", user.Username)
	}
	if user.ShadowsocksPassword != "ss-pass" {
		t.Fatalf("ShadowsocksPassword = %q, want ss-pass", user.ShadowsocksPassword)
	}
	if user.DataLimit != 5678 {
		t.Fatalf("DataLimit = %d, want 5678", user.DataLimit)
	}
	if user.DataLimitResetStrategy != "WEEK" {
		t.Fatalf("DataLimitResetStrategy = %q, want WEEK", user.DataLimitResetStrategy)
	}
}
