package gmgo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GetOauthToken(cfg *Config) (*oauth2.Token, error) {
	b, err := os.ReadFile(cfg.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("error reading credentials from %s: %w", cfg.CredentialsFile, err)
	}

	config, err := google.ConfigFromJSON(b, cfg.Scopes...)
	if err != nil {
		return nil, fmt.Errorf("error configuring OAuth2: %w", err)
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following URL and authorize the application:\n%v\n\n", authURL)
	fmt.Print("Enter the authorization code: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		panic(fmt.Sprintf("Error reading oauth code: %v", err))
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		panic(fmt.Sprintf("Error obtaining oauth token: %v", err))
	}

	saveToken(cfg.TokenFile, tok)

	return tok, nil
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving token to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(fmt.Sprintf("Error saving token: %v", err))
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
