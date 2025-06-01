package gmgo

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	CredentialsFile string   `json:"credentials_file"`
	TokenFile       string   `json:"token_file"`
	Scopes          []string `json:"scopes"`
}

func validateConfig(cfg *Config) error {
	if cfg.CredentialsFile == "" {
		return fmt.Errorf("credentials_file is required")
	}
	if cfg.TokenFile == "" {
		return fmt.Errorf("token_file is required")
	}

	if _, err := os.Stat(cfg.CredentialsFile); os.IsNotExist(err) {
		return fmt.Errorf("credentials file does not exist: %s", cfg.CredentialsFile)
	}

	if _, err := os.Stat(cfg.TokenFile); os.IsNotExist(err) {
		return fmt.Errorf("token file does not exist: %s", cfg.TokenFile)
	}

	return nil
}

func loadConfig(cfg *Config) (*oauth2.Config, *oauth2.Token, error) {
	b, err := os.ReadFile(cfg.CredentialsFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading credentials from %s: %w", cfg.CredentialsFile, err)
	}

	token, err := tokenFromFile(cfg.TokenFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading token from %s: %w", cfg.TokenFile, err)
	}

	oauthConfig, err := google.ConfigFromJSON(b, cfg.Scopes...)
	if err != nil {
		return nil, nil, fmt.Errorf("error configuring OAuth2: %w", err)
	}

	return oauthConfig, token, nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}
