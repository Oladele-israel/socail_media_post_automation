package token

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userID, email string) (string, string, error) {
	hours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))

	id := uuid.NewString() // generate the ID first so we can return it

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(hours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        id, // use the variable, not uuid.NewString() inline
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := t.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign token: %w", err) // 3 return values on error too
	}

	return signed, id, nil // return both the token string AND the id
}

func GenerateRefreshToken() string {
	return uuid.NewString()
}

func ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// Verify the signing method is what we expect — prevents algorithm switching attacks
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// pkg/token/jwt.go

// // GenerateAccessToken now returns the tokenID too — needed for blacklisting on logout
// func GenerateAccessToken(userID, email string) (tokenString, tokenID string, err error) {
//     hours, _ := strconv.Atoi(os.Getenv("JWT_EXPIRY_HOURS"))
//     id := uuid.NewString() // This is the JWT's unique ID (jti claim)

//     claims := Claims{
//         UserID: userID,
//         Email:  email,
//         RegisteredClaims: jwt.RegisteredClaims{
//             ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(hours) * time.Hour)),
//             IssuedAt:  jwt.NewNumericDate(time.Now()),
//             ID:        id, // jti — unique identifier for this specific token
//         },
//     }

//     t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//     signed, err := t.SignedString([]byte(os.Getenv("JWT_SECRET")))
//     if err != nil {
//         return "", "", fmt.Errorf("failed to sign token: %w", err)
//     }

//     return signed, id, nil // return both the token string AND the ID
// }
