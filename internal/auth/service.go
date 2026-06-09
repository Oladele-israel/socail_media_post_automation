// internal/auth/service.go
package auth

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Oladele-israel/socialmedia-post-automation/internal/auth/providers"
	"github.com/Oladele-israel/socialmedia-post-automation/pkg/cache"
	"github.com/Oladele-israel/socialmedia-post-automation/pkg/crypto"
	"github.com/Oladele-israel/socialmedia-post-automation/pkg/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo     *Repository
	cache    *cache.RedisClient // swap: was DB refresh tokens, now Redis
	registry *providers.Registry
}

func NewService(repo *Repository, cache *cache.RedisClient) *Service {
	return &Service{
		repo:     repo,
		cache:    cache,
		registry: providers.NewRegistry(),
	}
}

type AuthResponse struct {
	User          *User  `json:"user"`
	AccessToken   string `json:"access_token"`
	AccessTokenID string `json:"access_token_id"` // ← added, needed for blacklisting on logout
	RefreshToken  string `json:"refresh_token"`
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, input.Email, string(hashed), input.FullName)
	if err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *Service) RefreshTokens(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Look up userID from Redis
	userID, err := s.cache.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Rotate — delete old token immediately (prevents reuse)
	if err := s.cache.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserById(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.generateAuthResponse(ctx, user)
}

func (s *Service) Logout(ctx context.Context, refreshToken, accessTokenID string, accessTokenTTL time.Duration) error {
	// 1. Delete refresh token from Redis
	if err := s.cache.DeleteRefreshToken(ctx, refreshToken); err != nil {
		return err
	}

	// 2. Blacklist the access token so it can't be used even before it expires
	// TTL = remaining lifetime of the access token
	return s.cache.BlacklistAccessToken(ctx, accessTokenID, accessTokenTTL)
}

func (s *Service) generateAuthResponse(ctx context.Context, user *User) (*AuthResponse, error) {
	accessToken, tokenID, err := token.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken := token.GenerateRefreshToken()

	days, _ := strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_DAYS"))
	ttl := time.Duration(days) * 24 * time.Hour

	if err := s.cache.SetRefreshToken(ctx, refreshToken, user.ID, ttl); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &AuthResponse{
		User:          user,
		AccessToken:   accessToken,
		AccessTokenID: tokenID, // ← now tokenID is used, compiler is happy
		RefreshToken:  refreshToken,
	}, nil
}

// social methods in service.go — notice NO platform-specific code

func (s *Service) GenerateAuthURL(ctx context.Context, userID string, platformName providers.Platform) (string, error) {
	// Get the right provider from registry
	provider, err := s.registry.Get(platformName)
	if err != nil {
		return "", err // "unsupported platform: instagram"
	}

	state := uuid.NewString()
	// Store state with platform info so callback knows which provider to use
	stateValue := fmt.Sprintf("%s:%s", platformName, userID)
	if err := s.cache.SetOAuthState(ctx, state, stateValue, 10*time.Minute); err != nil {
		return "", fmt.Errorf("failed to store oauth state: %w", err)
	}

	return provider.GetAuthURL(state), nil
}

func (s *Service) HandleCallback(ctx context.Context, code, state string) error {
	// Retrieve stored state value
	stateValue, err := s.cache.GetOAuthState(ctx, state)
	if err != nil {
		return fmt.Errorf("invalid or expired oauth state")
	}
	_ = s.cache.DeleteOAuthState(ctx, state)

	// Parse "platform:userID" from state
	parts := strings.SplitN(stateValue, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid state format")
	}
	platformName := providers.Platform(parts[0])
	userID := parts[1]

	// Get the right provider — same code handles LinkedIn, Instagram, X
	provider, err := s.registry.Get(platformName)
	if err != nil {
		return err
	}

	tokenResp, err := provider.ExchangeCode(code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	profile, err := provider.FetchProfile(tokenResp.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to fetch profile: %w", err)
	}

	encryptedToken, err := crypto.Encrypt(tokenResp.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt token: %w", err)
	}

	account := &SocialAccount{
		UserID:         userID,
		Platform:       platformName,
		PlatformUserID: profile.PlatformUserID,
		Name:           profile.Name,
		Email:          profile.Email,
		AccessToken:    encryptedToken,
		TokenExpiresAt: time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return s.repo.SaveSocialAccount(ctx, account)
}

func (s *Service) GetSocialProfile(ctx context.Context, userID string, platform providers.Platform) (*SocialAccount, error) {
	return s.repo.GetSocialAccount(ctx, userID, platform)
}

func (s *Service) GetAllSocialProfiles(ctx context.Context, userID string) ([]*SocialAccount, error) {
	return s.repo.GetAllSocialAccounts(ctx, userID)
}

func (s *Service) DisconnectSocial(ctx context.Context, userID string, platform providers.Platform) error {
	return s.repo.DeleteSocialAccount(ctx, userID, platform)
}

func (s *Service) GetSocialToken(ctx context.Context, userID string, platform providers.Platform) (string, error) {
	account, err := s.repo.GetSocialAccount(ctx, userID, platform)
	if err != nil {
		return "", err
	}

	if time.Now().After(account.TokenExpiresAt) {
		return "", fmt.Errorf("%s token expired, please reconnect", platform)
	}

	return crypto.Decrypt(account.AccessToken)
}
