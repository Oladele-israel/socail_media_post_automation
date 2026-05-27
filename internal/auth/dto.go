// internal/auth/dto.go  ← create this new file, keep DTOs separate
package auth

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
