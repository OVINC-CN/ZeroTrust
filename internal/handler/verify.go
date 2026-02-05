package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ovinc/zerotrust/internal/config"
	"github.com/sirupsen/logrus"
)

type VerifyRequest struct {
	ClientIP  string `json:"client_ip"`
	SessionID string `json:"session_id"`
	Method    string `json:"method"`
	Protocol  string `json:"protocol"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer"`
	Accept    string `json:"accept"`
	TraceID   string `json:"trace_id"`
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
		Protocol:  r.Header.Get("X-Forwarded-Proto"),
		Host:      r.Header.Get("X-Forwarded-Host"),
		Path:      r.Header.Get("X-Forwarded-Uri"),
		UserAgent: r.Header.Get("User-Agent"),
		Referer:   r.Header.Get("Referer"),
		Accept:    r.Header.Get("Accept"),
		TraceID:   r.Header.Get(cfg.Auth.TraceIDHeader),
	}

	// get session id from cookies
	if cookie, err := r.Cookie(cfg.Auth.SessionCookieName); err == nil {
		req.SessionID = cookie.Value
	}

	// perform authentication
	doAuth(ctx, w, &req)
}
