package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Oladele-israel/socialmedia-post-automation/pkg/database"
)

// think if this as a DTO in nestjs
// The field name starts with a lowercase letter (password). In Go, that means the field is unexported (private) – code outside the auth package cannot see or access it directly.
type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	FullName  string    `db:"full_name" json:"full_name"`
	Password  string    `db:"password" json:"-"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Repository method – now accepts context
func (r *Repository) CreateUser(ctx context.Context, email, hashedPassword, fullName string) (*User, error) {
	var user User
	query := `
        INSERT INTO users (email, password, full_name)
        VALUES (:email, :password, :full_name)
        RETURNING id, email, full_name, created_at, updated_at
    `

	// Use NamedQueryContext instead of NamedQuery
	rows, err := r.db.NamedQueryContext(ctx, query, map[string]interface{}{
		"email":     email,
		"password":  hashedPassword,
		"full_name": fullName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.StructScan(&user); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
	}
	return &user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (r *Repository) GetUserById(ctx context.Context, id string) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}
