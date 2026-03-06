package bff

import (
	"os"
	"path/filepath"
)

const configDirName = ".a2a-playground"

// gitignoreContent is written to ~/.a2a-playground/.gitignore so preset data is not committed.
const gitignoreContent = "*\n"

// ConfigBaseDir returns the config base directory (~/.a2a-playground), creating it and
// the .gitignore file if missing.
func ConfigBaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	base := filepath.Join(home, configDirName)
	if err := os.MkdirAll(base, 0700); err != nil {
		return "", err
	}
	gitignorePath := filepath.Join(base, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0600); err != nil {
			return "", err
		}
	}
	return base, nil
}

// OAuthDir returns ~/.a2a-playground/oauth, creating it if missing.
func OAuthDir() (string, error) {
	base, err := ConfigBaseDir()
	if err != nil {
		return "", err
	}
	oauthDir := filepath.Join(base, "oauth")
	if err := os.MkdirAll(oauthDir, 0700); err != nil {
		return "", err
	}
	return oauthDir, nil
}
