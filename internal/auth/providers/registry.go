// internal/auth/providers/registry.go
package providers

import "fmt"

// Registry holds all available OAuth providers
// Adding a new platform = register it here, nothing else changes
type Registry struct {
	providers map[Platform]OAuthProvider
}

func NewRegistry() *Registry {
	r := &Registry{
		providers: make(map[Platform]OAuthProvider),
	}

	// Register all providers here
	r.Register(NewLinkedInProvider())
	// r.Register(NewInstagramProvider())  ← uncomment when ready
	// r.Register(NewXProvider())          ← uncomment when ready

	return r
}

func (r *Registry) Register(p OAuthProvider) {
	r.providers[p.Name()] = p
}

// Get returns the provider for a given platform
func (r *Registry) Get(platform Platform) (OAuthProvider, error) {
	p, ok := r.providers[platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
	return p, nil
}

// Available returns all registered platform names
func (r *Registry) Available() []Platform {
	platforms := make([]Platform, 0, len(r.providers))
	for p := range r.providers {
		platforms = append(platforms, p)
	}
	return platforms
}
