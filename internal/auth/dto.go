// internal/auth/dto.go  ← create this new file, keep DTOs separate
package auth

import (
	"time"

	"github.com/Oladele-israel/socialmedia-post-automation/internal/auth/providers"
)

// RegisterInput — all fields required with specific rules
type RegisterInput struct {
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=8,max=72"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
}

// LoginInput
type LoginInput struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshInput
type RefreshInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutInput
type LogoutInput struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// validate:"required"              // field must be present and non-zero
// validate:"required,email"        // must be present + valid email format
// validate:"required,min=8"        // string min 8 characters
// validate:"required,min=8,max=72" // string between 8-72 characters
// validate:"required,url"          // must be valid URL
// validate:"required,oneof=draft scheduled published" // enum values
// validate:"omitempty,min=2"       // optional but if present must be min 2
// validate:"required,uuid4"        // must be valid UUID v4
// validate:"required,gt=0"         // number greater than 0
// validate:"required,gte=0,lte=5"  // number between 0 and 5

// on struct handles all platform
type SocialAccount struct {
	ID             string             `db:"id"               json:"id"`
	UserID         string             `db:"user_id"           json:"user_id"`
	Platform       providers.Platform `db:"platform"          json:"platform"`
	PlatformUserID string             `db:"platform_user_id"  json:"platform_user_id"`
	Name           string             `db:"name"              json:"name"`
	Email          string             `db:"email"             json:"email"`
	AccessToken    string             `db:"access_token"      json:"-"`
	TokenExpiresAt time.Time          `db:"token_expires_at"  json:"token_expires_at"`
	CreatedAt      time.Time          `db:"created_at"        json:"created_at"`
	UpdatedAt      time.Time          `db:"updated_at"        json:"updated_at"`
}
