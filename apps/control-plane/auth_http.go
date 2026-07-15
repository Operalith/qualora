package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

func (a *App) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	required, err := a.setupRequired(r.Context())
	if err != nil {
		a.logger.Error("setup status failed", "error", err)
		writeError(w, http.StatusInternalServerError, "setup_status_failed", "setup status could not be loaded")
		return
	}
	writeJSON(w, http.StatusOK, SetupStatusResponse{SetupRequired: required, Version: QualoraVersion})
}

func (a *App) handleSetupAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	required, err := a.setupRequired(r.Context())
	if err != nil {
		a.logger.Error("setup admin status failed", "error", err)
		writeError(w, http.StatusInternalServerError, "setup_status_failed", "setup status could not be loaded")
		return
	}
	if !required {
		writeError(w, http.StatusConflict, "setup_complete", "first-run setup has already been completed")
		return
	}

	input, err := decodeSetupAdminRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input, err = normalizeSetupAdminRequest(input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_admin", err.Error())
		return
	}
	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		a.logger.Error("hash first admin password failed", "error", err)
		writeError(w, http.StatusInternalServerError, "password_hash_failed", "admin password could not be stored")
		return
	}
	user, err := a.store.CreateFirstLocalAdmin(r.Context(), input, passwordHash)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			writeError(w, http.StatusConflict, "setup_complete", "first-run setup has already been completed")
			return
		}
		a.logger.Error("create first admin failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_admin_failed", "first admin could not be created")
		return
	}
	response, err := a.createSessionResponse(w, r, *user, "completed")
	if err != nil {
		a.logger.Error("create first admin session failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_session_failed", "session could not be created")
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	input, err := decodeLoginRequest(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	input = normalizeLoginRequest(input)
	user, err := a.store.GetLocalUserByEmail(r.Context(), input.Email)
	if err != nil || user == nil || !user.IsActive || !verifyPassword(input.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is invalid")
		return
	}
	if err := a.store.MarkLocalUserLogin(r.Context(), user.ID); err != nil {
		a.logger.Error("mark local user login failed", "error", err)
		writeError(w, http.StatusInternalServerError, "login_failed", "login could not be completed")
		return
	}
	freshUser, err := a.store.GetLocalUserByEmail(r.Context(), user.Email)
	if err == nil && freshUser != nil {
		user = freshUser
	}
	response, err := a.createSessionResponse(w, r, *user, "")
	if err != nil {
		a.logger.Error("create login session failed", "error", err)
		writeError(w, http.StatusInternalServerError, "create_session_failed", "session could not be created")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	if session, _, err := a.sessionFromRequest(r); err == nil && session != nil {
		if err := a.store.RevokeUserSession(r.Context(), session.ID); err != nil {
			a.logger.Error("revoke user session failed", "error", err)
			writeError(w, http.StatusInternalServerError, "logout_failed", "logout could not be completed")
			return
		}
	}
	a.clearAuthCookies(w)
	writeJSON(w, http.StatusOK, map[string]any{"logged_out": true})
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method is not allowed")
		return
	}
	session, user, err := a.sessionFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusOK, MeResponse{Authenticated: false})
		return
	}
	_ = a.store.TouchUserSession(r.Context(), session.ID)
	sanitized := sanitizeUser(*user)
	writeJSON(w, http.StatusOK, MeResponse{Authenticated: true, User: &sanitized, ExpiresAt: &session.ExpiresAt})
}

func (a *App) withAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.auth.AuthDisabled || isPublicAPIPath(r.URL.Path) || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		session, user, err := a.sessionFromRequest(r)
		if err != nil {
			a.clearAuthCookies(w)
			writeError(w, http.StatusUnauthorized, "unauthenticated", "authentication is required")
			return
		}
		if requiresCSRF(r) && !a.validCSRF(r, session) {
			writeError(w, http.StatusForbidden, "csrf_required", "valid CSRF token is required")
			return
		}
		_ = a.store.TouchUserSession(r.Context(), session.ID)
		next.ServeHTTP(w, r.WithContext(withAuthenticatedContext(r.Context(), user, session)))
	})
}

func isPublicAPIPath(path string) bool {
	switch path {
	case "/healthz", "/api/v1/setup/status", "/api/v1/setup/admin", "/api/v1/auth/login", "/api/v1/auth/logout", "/api/v1/auth/me":
		return true
	default:
		return false
	}
}

func requiresCSRF(r *http.Request) bool {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func (a *App) validCSRF(r *http.Request, session *UserSession) bool {
	header := strings.TrimSpace(r.Header.Get(csrfHeaderName))
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || header == "" || cookie.Value == "" {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(header), []byte(cookie.Value)) != 1 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(hashToken(header)), []byte(session.CSRFTokenHash)) == 1
}

func (a *App) sessionFromRequest(r *http.Request) (*UserSession, *LocalUser, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return nil, nil, ErrNotFound
	}
	return a.store.GetActiveSessionByTokenHash(r.Context(), hashToken(cookie.Value), time.Now().UTC())
}

func (a *App) setupRequired(ctx context.Context) (bool, error) {
	if a.auth.AuthDisabled {
		return false, nil
	}
	count, err := a.store.CountLocalUsers(ctx)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (a *App) createSessionResponse(w http.ResponseWriter, r *http.Request, user LocalUser, setupStatus string) (AuthResponse, error) {
	sessionToken, err := randomToken(sessionTokenBytes)
	if err != nil {
		return AuthResponse{}, err
	}
	csrfToken, err := randomToken(csrfTokenBytes)
	if err != nil {
		return AuthResponse{}, err
	}
	expiresAt := time.Now().UTC().Add(a.auth.SessionTTL)
	session, err := a.store.CreateUserSession(r.Context(), user.ID, hashToken(sessionToken), hashToken(csrfToken), truncateForStorage(r.UserAgent(), 512), clientIP(r), expiresAt)
	if err != nil {
		return AuthResponse{}, err
	}
	a.setAuthCookies(w, sessionToken, csrfToken, session.ExpiresAt)
	return AuthResponse{User: sanitizeUser(user), ExpiresAt: session.ExpiresAt, CSRFToken: csrfToken, SetupStatus: setupStatus}, nil
}

func (a *App) setAuthCookies(w http.ResponseWriter, sessionToken string, csrfToken string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   a.auth.CookieSecure,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrfToken,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   maxAge,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   a.auth.CookieSecure,
	})
}

func (a *App) clearAuthCookies(w http.ResponseWriter) {
	expired := time.Unix(0, 0)
	for _, cookie := range []http.Cookie{
		{Name: sessionCookieName, HttpOnly: true},
		{Name: csrfCookieName, HttpOnly: false},
	} {
		cookie.Value = ""
		cookie.Path = "/"
		cookie.Expires = expired
		cookie.MaxAge = -1
		cookie.SameSite = http.SameSiteLaxMode
		cookie.Secure = a.auth.CookieSecure
		http.SetCookie(w, &cookie)
	}
}

func decodeSetupAdminRequest(w http.ResponseWriter, r *http.Request) (SetupAdminRequest, error) {
	var input SetupAdminRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, errors.New("request body must be valid setup admin JSON")
	}
	return input, nil
}

func decodeLoginRequest(w http.ResponseWriter, r *http.Request) (LoginRequest, error) {
	var input LoginRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return input, errors.New("request body must be valid login JSON")
	}
	return input, nil
}

func truncateForStorage(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	return value[:max]
}
