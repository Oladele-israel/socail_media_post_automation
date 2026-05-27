// internal/auth/service.go
package auth

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Oladele-israel/socialmedia-post-automation/pkg/cache"
	"github.com/Oladele-israel/socialmedia-post-automation/pkg/token"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo  *Repository
	cache *cache.RedisClient // swap: was DB refresh tokens, now Redis
}

func NewService(repo *Repository, cache *cache.RedisClient) *Service {
	return &Service{repo: repo, cache: cache}
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
