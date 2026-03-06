package bff

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
)

func getNicknameFromRequest(r *http.Request) string {
	return mux.Vars(r)["nickname"]
}

// ListOAuthPresetsRequest: no body for GET.
// ListOAuthPresetsResponse is the response for GET /api/oauth-presets.
type ListOAuthPresetsResponse struct {
	Presets []OAuthPresetSummary `json:"presets"`
}

// GetOAuthPresetResponse is the response for GET /api/oauth-presets/{nickname}.
type GetOAuthPresetResponse struct {
	Nickname         string `json:"nickname"`
	ClientID         string `json:"client_id"`
	ClientSecret     string `json:"client_secret"`
	AuthorizationURL string `json:"authorization_url"`
	TokenURL         string `json:"token_url"`
	Scope            string `json:"scope"`
}

// SaveOAuthPresetRequest is the request body for POST /api/oauth-presets.
type SaveOAuthPresetRequest struct {
	Nickname         string `json:"nickname"`
	ClientID         string `json:"client_id"`
	ClientSecret     string `json:"client_secret"`
	AuthorizationURL string `json:"authorization_url"`
	TokenURL         string `json:"token_url"`
	Scope            string `json:"scope"`
	Overwrite        bool   `json:"overwrite"`
}

// HandleListOAuthPresets handles GET /api/oauth-presets.
func HandleListOAuthPresets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	presets, err := ListOAuthPresets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ListOAuthPresetsResponse{Presets: presets})
}

// HandleGetOAuthPreset handles GET /api/oauth-presets/{nickname}.
func HandleGetOAuthPreset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	nickname := getNicknameFromRequest(r)
	if nickname == "" {
		http.Error(w, "nickname required", http.StatusBadRequest)
		return
	}
	if !ValidNickname(nickname) {
		http.Error(w, "invalid nickname: only letters, numbers, underscore, dot, and hyphen allowed", http.StatusBadRequest)
		return
	}
	preset, err := LoadOAuthPreset(nickname)
	if err != nil {
		if errors.Is(err, ErrPresetNotFound) {
			http.Error(w, "preset not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(GetOAuthPresetResponse{
		Nickname:         preset.Nickname,
		ClientID:         preset.ClientID,
		ClientSecret:     preset.ClientSecret,
		AuthorizationURL: preset.AuthorizationURL,
		TokenURL:         preset.TokenURL,
		Scope:            preset.Scope,
	})
}

// HandleSaveOAuthPreset handles POST /api/oauth-presets.
func HandleSaveOAuthPreset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req SaveOAuthPresetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Nickname == "" {
		http.Error(w, "nickname required", http.StatusBadRequest)
		return
	}
	if !ValidNickname(req.Nickname) {
		http.Error(w, "invalid nickname: only letters, numbers, underscore, dot, and hyphen allowed", http.StatusBadRequest)
		return
	}
	preset := OAuthPreset{
		Nickname:         req.Nickname,
		ClientID:         req.ClientID,
		ClientSecret:     req.ClientSecret,
		AuthorizationURL: req.AuthorizationURL,
		TokenURL:         req.TokenURL,
		Scope:            req.Scope,
	}
	err := SaveOAuthPreset(preset, req.Overwrite)
	if err != nil {
		if errors.Is(err, ErrPresetExists) {
			http.Error(w, "preset already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandleDeleteOAuthPreset handles DELETE /api/oauth-presets/{nickname}.
func HandleDeleteOAuthPreset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	nickname := getNicknameFromRequest(r)
	if nickname == "" {
		http.Error(w, "nickname required", http.StatusBadRequest)
		return
	}
	if !ValidNickname(nickname) {
		http.Error(w, "invalid nickname: only letters, numbers, underscore, dot, and hyphen allowed", http.StatusBadRequest)
		return
	}
	err := DeleteOAuthPreset(nickname)
	if err != nil {
		if errors.Is(err, ErrPresetNotFound) {
			http.Error(w, "preset not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
