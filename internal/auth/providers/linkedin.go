// internal/auth/providers/linkedin.go
package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// LinkedInProvider implements OAuthProvider
type LinkedInProvider struct{}

func NewLinkedInProvider() *LinkedInProvider {
	return &LinkedInProvider{}
}

// Name satisfies the OAuthProvider interface
func (l *LinkedInProvider) Name() Platform {
	return PlatformLinkedIn
}

// GetAuthURL builds the LinkedIn OAuth authorization URL
func (l *LinkedInProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", os.Getenv("LINKEDIN_CLIENT_ID"))
	params.Set("redirect_uri", os.Getenv("LINKEDIN_REDIRECT_URL"))
	params.Set("state", state)
	params.Set("scope", os.Getenv("LINKEDIN_SCOPES"))

	return "https://www.linkedin.com/oauth/v2/authorization?" + params.Encode()
}

// ExchangeCode swaps the auth code for an access token
func (l *LinkedInProvider) ExchangeCode(code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", os.Getenv("LINKEDIN_REDIRECT_URL"))
	data.Set("client_id", os.Getenv("LINKEDIN_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("LINKEDIN_CLIENT_SECRET"))

	resp, err := http.PostForm(
		"https://www.linkedin.com/oauth/v2/accessToken",
		data,
	)
	if err != nil {
		return nil, fmt.Errorf("linkedin token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkedin token error: %s", string(body))
	}

	// LinkedIn specific response shape
	var raw struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Normalize into our shared TokenResponse
	return &TokenResponse{
		AccessToken: raw.AccessToken,
		ExpiresIn:   raw.ExpiresIn,
	}, nil
}

// FetchProfile gets the LinkedIn user's profile
func (l *LinkedInProvider) FetchProfile(accessToken string) (*Profile, error) {
	req, err := http.NewRequest("GET", "https://api.linkedin.com/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("linkedin profile request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkedin profile error: %s", string(body))
	}

	// LinkedIn specific profile shape
	var raw struct {
		Sub     string `json:"sub"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	// Normalize into our shared Profile
	return &Profile{
		PlatformUserID: raw.Sub,
		Name:           raw.Name,
		Email:          raw.Email,
		Picture:        raw.Picture,
	}, nil
}
