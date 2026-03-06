package bff

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
)

// Nickname must be letters, numbers, underscore, dot, or hyphen only.
var nicknameRegex = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// ErrPresetExists is returned by SaveOAuthPreset when overwrite is false and the file exists.
var ErrPresetExists = errors.New("preset already exists")

// ErrPresetNotFound is returned when loading or deleting a non-existent preset.
var ErrPresetNotFound = errors.New("preset not found")

// OAuthPreset is the persisted OAuth config (snake_case for JSON API).
type OAuthPreset struct {
	Nickname         string `json:"nickname"`
	ClientID         string `json:"client_id"`
	ClientSecret     string `json:"client_secret"`
	AuthorizationURL string `json:"authorization_url"`
	TokenURL         string `json:"token_url"`
	Scope            string `json:"scope"`
}

// OAuthPresetSummary is returned in the list response.
type OAuthPresetSummary struct {
	Nickname string `json:"nickname"`
}

// ValidNickname returns true if nickname matches ^[A-Za-z0-9_.-]+$.
func ValidNickname(nickname string) bool {
	return nickname != "" && nicknameRegex.MatchString(nickname)
}

// presetPath returns the file path for a preset (caller must ensure nickname is valid).
func presetPath(nickname string) (string, error) {
	dir, err := OAuthDir()
	if err != nil {
		return "", err
	}
	// Sanitize: only use nickname as filename base (no path traversal)
	return filepath.Join(dir, filepath.Clean(nickname)+".json"), nil
}

// ListOAuthPresets returns all preset nicknames in ~/.a2a-playground/oauth.
func ListOAuthPresets() ([]OAuthPresetSummary, error) {
	dir, err := OAuthDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []OAuthPresetSummary
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}
		nickname := name[:len(name)-len(".json")]
		if !ValidNickname(nickname) {
			continue
		}
		out = append(out, OAuthPresetSummary{Nickname: nickname})
	}
	return out, nil
}

// LoadOAuthPreset reads a preset by nickname. Returns error if nickname invalid or file missing.
func LoadOAuthPreset(nickname string) (OAuthPreset, error) {
	if !ValidNickname(nickname) {
		return OAuthPreset{}, errors.New("invalid nickname")
	}
	path, err := presetPath(nickname)
	if err != nil {
		return OAuthPreset{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return OAuthPreset{}, ErrPresetNotFound
		}
		return OAuthPreset{}, err
	}
	var p OAuthPreset
	if err := json.Unmarshal(data, &p); err != nil {
		return OAuthPreset{}, err
	}
	p.Nickname = nickname
	return p, nil
}

// SaveOAuthPreset writes a preset. If overwrite is false and the file exists, returns ErrPresetExists.
func SaveOAuthPreset(p OAuthPreset, overwrite bool) error {
	if !ValidNickname(p.Nickname) {
		return errors.New("invalid nickname")
	}
	path, err := presetPath(p.Nickname)
	if err != nil {
		return err
	}
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return ErrPresetExists
		}
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// DeleteOAuthPreset removes a preset file. Returns error if nickname invalid or file not found.
func DeleteOAuthPreset(nickname string) error {
	if !ValidNickname(nickname) {
		return errors.New("invalid nickname")
	}
	path, err := presetPath(nickname)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrPresetNotFound
		}
		return err
	}
	return nil
}
