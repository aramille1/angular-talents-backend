package domain

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// AdminJwtCustomClaims defines custom claims for admin JWT tokens
type AdminJwtCustomClaims struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	jwt.StandardClaims
}

// GenerateAdminJWT generates a JWT token for an admin
func GenerateAdminJWT(adminID string) (string, error) {
	// Ensure JWT secret is available
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT secret is not set")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &AdminJwtCustomClaims{
		ID:   adminID,
		Role: "admin",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateAdminToken validates an admin JWT token
func ValidateAdminToken(signedToken string) (string, string, error) {
	// Ensure JWT secret is available
	if len(jwtSecret) == 0 {
		return "", "", errors.New("JWT secret is not set")
	}

	token, err := jwt.ParseWithClaims(
		signedToken,
		&AdminJwtCustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		},
	)

	if err != nil {
		return "", "", err
	}

	if !token.Valid {
		return "", "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*AdminJwtCustomClaims)
	if !ok {
		return "", "", errors.New("failed to extract claims")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return "", "", errors.New("token expired")
	}

	return claims.ID, claims.Role, nil
}
