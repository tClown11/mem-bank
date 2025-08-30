package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken  = errors.New("invalid token")
	ErrExpiredToken  = errors.New("token has expired")
	ErrTokenNotFound = errors.New("token not found")
)

// Claims represents the JWT claims
type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey []byte
	issuer    string
	expiry    time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, issuer string, expiry time.Duration) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		expiry:    expiry,
	}
}

// GenerateToken generates a new JWT token for the user
func (j *JWTService) GenerateToken(userID uuid.UUID, username, email, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiry
func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		// Allow refresh of expired tokens (within reasonable time)
		if !errors.Is(err, ErrExpiredToken) {
			return "", err
		}

		// Parse expired token to get claims
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return j.secretKey, nil
		})
		if err != nil {
			return "", ErrInvalidToken
		}

		var ok bool
		claims, ok = token.Claims.(*Claims)
		if !ok {
			return "", ErrInvalidToken
		}

		// Check if token is too old to refresh (e.g., more than 7 days)
		if time.Since(claims.ExpiresAt.Time) > 7*24*time.Hour {
			return "", ErrExpiredToken
		}
	}

	// Generate new token with same user info
	return j.GenerateToken(claims.UserID, claims.Username, claims.Email, claims.Role)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
	const bearerPrefix = "Bearer "
	if len(authHeader) > len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		return authHeader[len(bearerPrefix):]
	}
	return ""
}
