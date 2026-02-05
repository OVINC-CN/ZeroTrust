package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ovinc/zerotrust/internal/config"
	"github.com/ovinc/zerotrust/internal/session"
	"github.com/ovinc/zerotrust/internal/store"
	"github.com/sirupsen/logrus"
)

type VerifyRequest struct {
	ClientIP  string `json:"client_ip"`
	SessionID string `json:"session_id"`
	Method    string `json:"method"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer"`
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// only POST method is allowed
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// decode request body into struct
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.WithContext(ctx).Warnf("failed to decode request body: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// perform authentication
	doAuth(ctx, w, &req)
}

func ForwardAuthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cfg := config.Get()

	// load info from headers
	req := VerifyRequest{
		ClientIP:  r.Header.Get(cfg.Auth.ClientIPHeader),
		Method:    r.Header.Get("X-Forwarded-Method"),
		Host:      r.Header.Get("X-Forwarded-Host"),
		Path:      r.Header.Get("X-Forwarded-Uri"),
		UserAgent: r.Header.Get("User-Agent"),
		Referer:   r.Header.Get("Referer"),
	}

	// get session id from cookies
	if cookie, err := r.Cookie(cfg.Auth.SessionCookieName); err == nil {
		req.SessionID = cookie.Value
	}

	// perform authentication
	doAuth(ctx, w, &req)
}

func doAuth(ctx context.Context, w http.ResponseWriter, req *VerifyRequest) {
	// log incoming verify request
	logrus.WithContext(ctx).Infof(
		"verifying request from %s\n%s %s%s\nuser_agent: %s\nreferer: %s",
		req.ClientIP,
		req.Method,
		req.Host,
		req.Path,
		req.UserAgent,
		req.Referer,
	)

	// check if session id is provided
	if req.SessionID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// get session data from redis
	sessionData, err := store.GetSession(ctx, req.SessionID)
	if err != nil {
		logrus.WithContext(ctx).Warnf("failed to get session from store: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// parse django session to extract user info
	userInfo, err := session.ParseDjangoSession(ctx, []byte(sessionData))
	if err != nil {
		logrus.WithContext(ctx).Warnf("failed to parse session data: %v", err)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// log successful authorization
	logrus.WithContext(ctx).Infof("request authorized for %s (%s)", userInfo.UserID, maskSessionID(req.SessionID))

	// return success response with user id
	w.WriteHeader(http.StatusOK)
}

func maskSessionID(sessionID string) string {
	// return masked placeholder for short session ids
	if len(sessionID) <= 8 {
		return "****"
	}
	// mask middle part for security
	return sessionID[:4] + "****" + sessionID[len(sessionID)-4:]
}
