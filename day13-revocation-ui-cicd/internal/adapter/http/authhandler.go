package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/sitthichok-J/self-study-2026/day13-revocation-ui-cicd/internal/auth"
)

// CSRF defence for the cookie-authenticated routes (refresh, logout): the
// cookie is SameSite=Strict, and a request must also carry this header. A
// plain HTML form cannot add it, and a cross-site fetch that adds it needs a
// CORS preflight that the gateway only grants to the web origin.
const (
	CSRFHeader = "X-Requested-With"
	CSRFValue  = "claims-web"
)

// CookieConfig describes the refresh-token cookie.
type CookieConfig struct {
	Name   string // default "refresh_token"
	Path   string // path the BROWSER sees (behind the gateway: "/api/auth"); default "/api/auth"
	Secure bool   // sent over https only. Zero value (false) is only for tests: the API config defaults it to true
}

func (c CookieConfig) withDefaults() CookieConfig {
	if c.Name == "" {
		c.Name = "refresh_token"
	}
	if c.Path == "" {
		c.Path = "/api/auth"
	}
	return c
}

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (h *Handler) setRefreshCookie(w nethttp.ResponseWriter, t auth.Tokens) {
	nethttp.SetCookie(w, &nethttp.Cookie{
		Name: h.cookie.Name, Value: t.Refresh, Path: h.cookie.Path,
		Expires:  t.RefreshExpires,
		MaxAge:   t.RefreshMaxAge,
		HttpOnly: true, Secure: h.cookie.Secure, SameSite: nethttp.SameSiteStrictMode,
	})
}

func (h *Handler) clearRefreshCookie(w nethttp.ResponseWriter) {
	nethttp.SetCookie(w, &nethttp.Cookie{
		Name: h.cookie.Name, Value: "", Path: h.cookie.Path, MaxAge: -1,
		HttpOnly: true, Secure: h.cookie.Secure, SameSite: nethttp.SameSiteStrictMode,
	})
}

func (h *Handler) writeTokens(w nethttp.ResponseWriter, t auth.Tokens) {
	h.setRefreshCookie(w, t)
	writeJSON(w, nethttp.StatusOK, tokenResponse{AccessToken: t.Access, TokenType: "Bearer", ExpiresIn: t.ExpiresIn})
}

func (h *Handler) login(w nethttp.ResponseWriter, r *nethttp.Request) {
	r.Body = nethttp.MaxBytesReader(w, r.Body, 4<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var b loginBody
	if err := dec.Decode(&b); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, errBody{Error: "request body must be JSON with username and password"})
		return
	}
	t, err := h.sess.Login(r.Context(), b.Username, b.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		// One message for unknown user and wrong password. The reason stays in the log, never the username or password.
		h.log.WarnContext(r.Context(), "login failed", "security_event", "login_failed")
		writeJSON(w, nethttp.StatusUnauthorized, errBody{Error: "invalid username or password"})
		return
	}
	if err != nil {
		writeErr(w, r, h.log, err, "")
		return
	}
	h.writeTokens(w, t)
}

// requireCSRF answers 403 unless the anti-CSRF header is present.
func (h *Handler) requireCSRF(w nethttp.ResponseWriter, r *nethttp.Request) bool {
	if r.Header.Get(CSRFHeader) == CSRFValue {
		return true
	}
	h.log.WarnContext(r.Context(), "auth request without CSRF header", "security_event", "csrf_header_missing")
	writeJSON(w, nethttp.StatusForbidden, errBody{Error: "forbidden"})
	return false
}

func (h *Handler) refreshToken(r *nethttp.Request) string {
	c, err := r.Cookie(h.cookie.Name)
	if err != nil {
		return ""
	}
	return c.Value
}

func (h *Handler) refresh(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.requireCSRF(w, r) {
		return
	}
	t, err := h.sess.Refresh(r.Context(), h.refreshToken(r))
	switch {
	case err == nil:
		h.writeTokens(w, t)
	case errors.Is(err, auth.ErrInvalidRefresh):
		if errors.Is(err, auth.ErrRefreshReused) {
			h.log.WarnContext(r.Context(), "refresh token reused: session family revoked", "security_event", "refresh_reuse")
		} else {
			h.log.WarnContext(r.Context(), "refresh failed", "security_event", "refresh_failed")
		}
		h.clearRefreshCookie(w)
		writeJSON(w, nethttp.StatusUnauthorized, errBody{Error: "unauthorized"})
	default:
		writeErr(w, r, h.log, err, "")
	}
}

func (h *Handler) logout(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.requireCSRF(w, r) {
		return
	}
	if err := h.sess.Logout(r.Context(), h.refreshToken(r)); err != nil {
		writeErr(w, r, h.log, err, "")
		return
	}
	// The access token the browser still holds goes on the blocklist until it would expire anyway.
	// No Bearer header or an invalid one is fine: logging out twice, or with an expired token, stays harmless.
	if h.opt.Revoker != nil {
		if p, err := h.ver.Verify(auth.BearerToken(r.Header.Get("Authorization"))); err == nil {
			if err := h.opt.Revoker.RevokeToken(r.Context(), p.JTI, h.ver.RevokeUntil(p)); err != nil {
				h.log.ErrorContext(r.Context(), "token revoke failed", "security_event", "revoke_failed", "err", err)
				h.clearRefreshCookie(w) // refresh family is already revoked; tell the client to stop using it
				writeJSON(w, nethttp.StatusServiceUnavailable, errBody{Error: "logout incomplete, try again"})
				return
			}
			h.log.InfoContext(r.Context(), "access token revoked at logout", "security_event", "token_revoked")
		}
	}
	h.clearRefreshCookie(w)
	w.WriteHeader(nethttp.StatusNoContent)
}
