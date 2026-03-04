package bff

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const oauthStateTTL = 10 * time.Minute

type oauthPendingState struct {
	ClientID       string
	ClientSecret   string
	TokenURL       string
	RedirectOrigin string // optional; when set, callback redirects here (e.g. SPA origin for popup flow)
	CreatedAt      time.Time
}

var oauthStateStore = struct {
	sync.Mutex
	m map[string]oauthPendingState
}{m: make(map[string]oauthPendingState)}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func putOAuthState(state string, s oauthPendingState) {
	s.CreatedAt = time.Now()
	oauthStateStore.Lock()
	defer oauthStateStore.Unlock()
	oauthStateStore.m[state] = s
}

func getOAuthState(state string) (oauthPendingState, bool) {
	oauthStateStore.Lock()
	defer oauthStateStore.Unlock()
	s, ok := oauthStateStore.m[state]
	if !ok {
		return oauthPendingState{}, false
	}
	if time.Since(s.CreatedAt) > oauthStateTTL {
		delete(oauthStateStore.m, state)
		return oauthPendingState{}, false
	}
	delete(oauthStateStore.m, state)
	return s, true
}

// StartAuthRequest is the JSON body for POST /auth/start.
type StartAuthRequest struct {
	ClientID         string `json:"client_id"`
	ClientSecret     string `json:"client_secret"`
	AuthorizationURL string `json:"authorization_url"`
	TokenURL         string `json:"token_url"`
	Scope            string `json:"scope"`
	RedirectOrigin   string `json:"redirect_origin"` // optional; SPA origin for callback redirect (e.g. popup flow)
}

// StartAuthResponse is the JSON response from POST /auth/start.
type StartAuthResponse struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
}

// HandleAuthStart handles POST /auth/start: accepts client config, stores in state, returns auth URL.
func HandleAuthStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req StartAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.AuthorizationURL == "" || req.TokenURL == "" {
		http.Error(w, "authorization_url and token_url required", http.StatusBadRequest)
		return
	}
	state, err := generateState()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	redirectURI := getRequestOrigin(r) + "/auth/callback"
	putOAuthState(state, oauthPendingState{
		ClientID:       req.ClientID,
		ClientSecret:   req.ClientSecret,
		TokenURL:       req.TokenURL,
		RedirectOrigin: req.RedirectOrigin,
	})

	authURL := req.AuthorizationURL +
		"?response_type=code" +
		"&client_id=" + url.QueryEscape(req.ClientID) +
		"&redirect_uri=" + url.QueryEscape(redirectURI) +
		"&state=" + url.QueryEscape(state)
	if req.Scope != "" {
		authURL += "&scope=" + url.QueryEscape(req.Scope)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(StartAuthResponse{AuthURL: authURL, State: state})
}

func getRequestOrigin(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	host := r.Host
	if host == "" {
		host = "localhost:3000"
	}
	return scheme + "://" + host
}

// HandleAuthCallback handles GET /auth/callback: exchange code for tokens, redirect to SPA with tokens in fragment.
func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Error(w, "code and state required", http.StatusBadRequest)
		return
	}
	pending, ok := getOAuthState(state)
	if !ok {
		http.Error(w, "invalid or expired state", http.StatusBadRequest)
		return
	}
	redirectURI := getRequestOrigin(r) + "/auth/callback"
	tokens, err := ExchangeCode(pending.TokenURL, code, redirectURI, pending.ClientID, pending.ClientSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	origin := getRequestOrigin(r)
	if pending.RedirectOrigin != "" {
		origin = pending.RedirectOrigin
	}
	// Redirect to SPA with tokens in hash fragment so SPA can read and store them (no server round-trip)
	redirectTo := origin + "/#auth_callback?access_token=" + url.QueryEscape(tokens.AccessToken) +
		"&refresh_token=" + url.QueryEscape(tokens.RefreshToken) +
		"&expires_in=" + strconv.FormatInt(tokens.ExpiresIn, 10)
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

// RefreshRequest is the JSON body for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TokenURL     string `json:"token_url"`
}

// HandleAuthRefresh handles POST /auth/refresh: exchange refresh_token for new tokens.
func HandleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.RefreshToken == "" || req.TokenURL == "" {
		http.Error(w, "refresh_token and token_url required", http.StatusBadRequest)
		return
	}
	tokens, err := RefreshTokens(req.TokenURL, req.RefreshToken, req.ClientID, req.ClientSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokens)
}