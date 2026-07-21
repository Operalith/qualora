package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	QualoraVersion     = "0.22.0-alpha"
	sessionCookieName  = "qualora_session"
	csrfCookieName     = "qualora_csrf"
	csrfHeaderName     = "X-Qualora-CSRF"
	minPasswordLength  = 12
	sessionTokenBytes  = 32
	csrfTokenBytes     = 32
	passwordSaltBytes  = 16
	passwordHashBytes  = 32
	argon2MemoryKiB    = 64 * 1024
	argon2Iterations   = 1
	argon2Parallelism  = 4
	maxAuthBodyBytes   = 64 * 1024
	authContextUserKey = "qualora.auth.user"
	authContextSessKey = "qualora.auth.session"
)

type AuthConfig struct {
	SessionTTL   time.Duration
	CookieSecure bool
	AuthDisabled bool
}

type authContextKey string

func sanitizeUser(user LocalUser) AuthUser {
	return AuthUser{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func normalizeSetupAdminRequest(input SetupAdminRequest) (SetupAdminRequest, error) {
	input.Email = normalizeEmail(input.Email)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.Email == "" || !strings.Contains(input.Email, "@") {
		return input, fmt.Errorf("email must be valid")
	}
	if len(input.DisplayName) < 2 || len(input.DisplayName) > 120 {
		return input, fmt.Errorf("display_name must be between 2 and 120 characters")
	}
	if input.ConfirmPassword != "" && input.Password != input.ConfirmPassword {
		return input, fmt.Errorf("password confirmation does not match")
	}
	if err := validatePassword(input.Password); err != nil {
		return input, err
	}
	return input, nil
}

func normalizeLoginRequest(input LoginRequest) LoginRequest {
	input.Email = normalizeEmail(input.Email)
	return input
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}
	weak := map[string]bool{
		"password":               true,
		"password123":            true,
		"qualora":                true,
		"qualora123":             true,
		"qualora-password":       true,
		"qualora-admin":          true,
		"qualora-admin-password": false,
	}
	if weak[strings.ToLower(strings.TrimSpace(password))] {
		return fmt.Errorf("password is too common")
	}
	return nil
}

func hashPassword(password string) (string, error) {
	salt, err := randomBytes(passwordSaltBytes)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2MemoryKiB, argon2Parallelism, passwordHashBytes)
	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2MemoryKiB,
		argon2Iterations,
		argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(password string, encoded string) bool {
	params, salt, expected, err := parsePasswordHash(encoded)
	if err != nil {
		return false
	}
	actual := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

type passwordHashParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
}

func parsePasswordHash(encoded string) (passwordHashParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return passwordHashParams{}, nil, nil, fmt.Errorf("unsupported password hash")
	}
	paramParts := strings.Split(parts[3], ",")
	if len(paramParts) != 3 {
		return passwordHashParams{}, nil, nil, fmt.Errorf("invalid password hash params")
	}
	params := passwordHashParams{}
	for _, part := range paramParts {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			return passwordHashParams{}, nil, nil, fmt.Errorf("invalid password hash param")
		}
		value, err := strconv.Atoi(keyValue[1])
		if err != nil || value <= 0 {
			return passwordHashParams{}, nil, nil, fmt.Errorf("invalid password hash param value")
		}
		switch keyValue[0] {
		case "m":
			params.memory = uint32(value)
		case "t":
			params.iterations = uint32(value)
		case "p":
			params.parallelism = uint8(value)
		default:
			return passwordHashParams{}, nil, nil, fmt.Errorf("unknown password hash param")
		}
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return passwordHashParams{}, nil, nil, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return passwordHashParams{}, nil, nil, err
	}
	return params, salt, hash, nil
}

func randomToken(size int) (string, error) {
	bytes, err := randomBytes(size)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func randomBytes(size int) ([]byte, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return nil, fmt.Errorf("generate random bytes: %w", err)
	}
	return value, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func authenticatedUser(r *http.Request) (*LocalUser, bool) {
	user, ok := r.Context().Value(authContextKey(authContextUserKey)).(*LocalUser)
	return user, ok && user != nil
}

func authenticatedSession(r *http.Request) (*UserSession, bool) {
	session, ok := r.Context().Value(authContextKey(authContextSessKey)).(*UserSession)
	return session, ok && session != nil
}

func withAuthenticatedContext(ctx context.Context, user *LocalUser, session *UserSession) context.Context {
	ctx = context.WithValue(ctx, authContextKey(authContextUserKey), user)
	return context.WithValue(ctx, authContextKey(authContextSessKey), session)
}
