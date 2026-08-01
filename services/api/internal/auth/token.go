package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type TokenClaims struct {
	UserID         string `json:"uid"`
	OrganizationID string `json:"oid"`
	SessionID      string `json:"sid"`
	MFA            bool   `json:"mfa"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret        []byte
	issuer        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewTokenManager(secret string) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("JWT secret must be at least 32 bytes")
	}

	return &TokenManager{
		secret:        []byte(secret),
		issuer:        "rtdigital-api",
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 30 * 24 * time.Hour,
	}, nil
}

func (tm *TokenManager) AccessExpiry() time.Duration {
	return tm.accessExpiry
}

func (tm *TokenManager) RefreshExpiry() time.Duration {
	return tm.refreshExpiry
}

func (tm *TokenManager) IssueAccessToken(claims TokenClaims) (string, error) {
	now := time.Now().UTC()
	claims.Issuer = tm.issuer
	claims.IssuedAt = jwt.NewNumericDate(now)
	claims.NotBefore = jwt.NewNumericDate(now)
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(tm.accessExpiry))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func (tm *TokenManager) VerifyAccessToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secret, nil
	}, jwt.WithIssuer(tm.issuer), jwt.WithLeeway(5*time.Second))
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid || claims.UserID == "" || claims.OrganizationID == "" || claims.SessionID == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func GenerateOpaqueToken() (raw, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("generate opaque token: %w", err)
	}

	raw = base64.RawURLEncoding.EncodeToString(bytes)
	return raw, HashOpaqueToken(raw), nil
}

func HashOpaqueToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}