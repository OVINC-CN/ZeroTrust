package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ovinc/zerotrust/internal/config"
	"github.com/ovinc/zerotrust/internal/session"
	"github.com/ovinc/zerotrust/internal/store"
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
		unauthorizedResponse(w, req)
		return
	}

	// get session data from redis
	sessionData, err := store.GetSession(ctx, req.SessionID)
	if err != nil {
		logrus.WithContext(ctx).Warnf("failed to get session from store: %v", err)
		unauthorizedResponse(w, req)
		return
	}

	// parse django session to extract user info
	userInfo, err := session.ParseDjangoSession(ctx, []byte(sessionData))
	if err != nil {
		logrus.WithContext(ctx).Warnf("failed to parse session data: %v", err)
		unauthorizedResponse(w, req)
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

const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>身份验证失败</title>
    <style>
        :root {
            --bg-color: #f8f9fa;
            --text-primary: #374151;
            --text-secondary: #6b7280;
            --accent-color: #9ca3af;
            --font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
        }

        body {
            background-color: var(--bg-color);
            font-family: var(--font-family);
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            padding: 20px;
            box-sizing: border-box;
        }

        .container {
            max-width: 400px;
            width: 100%;
            text-align: center;
        }

        .icon-box {
            margin-bottom: 20px;
        }

        .icon-box svg {
            width: 48px;
            height: 48px;
            color: var(--accent-color);
        }

        h1 {
            color: var(--text-primary);
            font-size: 20px;
            margin: 0 0 12px 0;
            font-weight: 600;
        }

        p {
            color: var(--text-secondary);
            font-size: 14px;
            line-height: 1.6;
            margin: 0;
        }

        .redirect-btn {
            display: ${url ? 'inline-block' : 'none'};
            margin-top: 24px;
            padding: 10px 24px;
            background-color: var(--text-primary);
            color: #fff;
            border: none;
            border-radius: 6px;
            font-size: 14px;
            font-family: var(--font-family);
            cursor: pointer;
            transition: background-color 0.2s;
        }

        .redirect-btn:hover {
            background-color: var(--text-secondary);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon-box">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
            </svg>
        </div>
        <h1>身份验证失败</h1>
        <p>无法验证您的身份信息，请登录后重试</p>
        <button class="redirect-btn" onclick="window.location.href='{{.url}}'">前往登录</button>
    </div>
</body>
</html>`

func unauthorizedResponse(w http.ResponseWriter, req *VerifyRequest) {
	cfg := config.Get()
	w.WriteHeader(http.StatusUnauthorized)

	// response html
	if strings.Contains(req.Accept, "text/html") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		htmlContent := strings.ReplaceAll(
			htmlTemplate,
			"{{.url}}",
			fmt.Sprintf(
				"%s?%s=%s",
				cfg.Auth.LoginUrl,
				cfg.Auth.LoginRedirectParam,
				url.QueryEscape(fmt.Sprintf("%s://%s%s", req.Protocol, req.Host, req.Path)),
			),
		)
		_, _ = w.Write([]byte(htmlContent))
		return
	}

	// response json
	w.Header().Set("Content-Type", "application/json")
	data := map[string]interface{}{"code": 401, "error": "unauthorized", "message": "unauthorized", "data": nil}
	_ = json.NewEncoder(w).Encode(data)
}
