package providers

// interface definition to ensure dynamic socials

// Platform is a string type — prevents passing arbitrary strings
// as platform names throughout the codebase
type Platform string

const (
	PlatformLinkedIn  Platform = "linkedin"
	PlatformInstagram Platform = "instagram"
	PlatformX         Platform = "x"
)

// TokenResponse is the normalized token data every provider returns
// Each provider's API returns different field names — we normalize here
type TokenResponse struct {
	AccessToken string
	ExpiresIn   int // seconds
}

// Profile is the normalized user profile every provider returns
type Profile struct {
	PlatformUserID string // the user's ID on that platform
	Name           string
	Email          string
	Picture        string
}

// OAuthProvider is the interface every social platform must implement
// Adding a new platform = create a new file that satisfies this interface
// Nothing else in the codebase changes
type OAuthProvider interface {
	// GetAuthURL builds the authorization URL to redirect the user to
	GetAuthURL(state string) string

	// ExchangeCode swaps the temporary code for an access token
	ExchangeCode(code string) (*TokenResponse, error)

	// FetchProfile gets the user's profile using their access token
	FetchProfile(accessToken string) (*Profile, error)

	// Name returns the platform identifier
	Name() Platform
}
