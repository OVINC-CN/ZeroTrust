package handler

import (
	"bytes"
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

func maskSessionID(sessionID string) string {
	// return masked placeholder for short session ids
	if len(sessionID) <= 8 {
		return "****"
	}
	// mask middle part for security
	return sessionID[:4] + "****" + sessionID[len(sessionID)-4:]
}

func unauthorizedResponse(ctx context.Context, w http.ResponseWriter, req *VerifyRequest) {
	cfg := config.Get()

	// response html
	if strings.Contains(req.Accept, "text/html") {
		// build data
		initData := map[string]interface{}{
			"url": fmt.Sprintf(
				"%s?%s=%s",
				cfg.Auth.LoginUrl,
				cfg.Auth.LoginRedirectParam,
				url.QueryEscape(fmt.Sprintf("%s://%s%s", req.Protocol, req.Host, req.Path)),
			),
			"urlDisplayStyle":     "inline-block",
			"traceID":             req.RequestID,
			"traceIdDisplayStyle": "block",
		}
		if req.RequestID == "" {
			initData["traceIdDisplayStyle"] = "none"
		}
		// parse template
		var buf bytes.Buffer
		if err := htmlTemplate.Execute(&buf, initData); err != nil {
			logrus.WithContext(ctx).WithError(err).Error("[UnauthorizedResponse] failed to execute html template")
			return
		}
		// write response
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(buf.Bytes())
		return
	}

	// response json
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	data := map[string]interface{}{"code": 401, "error": "unauthorized", "message": "unauthorized", "data": nil}
	_ = json.NewEncoder(w).Encode(data)
}

func doAuth(ctx context.Context, w http.ResponseWriter, req *VerifyRequest) {
	cfg := config.Get()

	var (
		result    string
		reason    string
		userID    string
		sessionID string
		logErr    error
	)
	defer func() {
		fields := logrus.Fields{
			"client_ip":  req.ClientIP,
			"method":     req.Method,
			"protocol":   req.Protocol,
			"host":       req.Host,
			"path":       req.Path,
			"request_id": req.RequestID,
			"user_agent": req.UserAgent,
			"referer":    req.Referer,
			"user_id":    "",
			"session_id": "",
			"reason":     "",
		}
		if userID != "" {
			fields["user_id"] = userID
		}
		if sessionID != "" {
			fields["session_id"] = maskSessionID(sessionID)
		}
		if reason != "" {
			fields["reason"] = reason
		}

		entry := logrus.WithContext(ctx).WithFields(fields)
		if logErr != nil {
			entry.WithError(logErr).Warn(result)
			return
		}
		entry.Info(result)
	}()

	// check methods
	skipVerify := true
	reqMethod := strings.ToLower(req.Method)
	for _, m := range cfg.Auth.VerifyMethods {
		if strings.ToLower(m) == reqMethod {
			skipVerify = false
			break
		}
	}
	if skipVerify {
		result = "request skipped"
		reason = "method_not_verified"
		w.WriteHeader(http.StatusOK)
		return
	}

	// check if session id is provided
	if req.SessionID == "" {
		result = "request unauthorized"
		reason = "missing_session"
		unauthorizedResponse(ctx, w, req)
		return
	}

	// get session data from redis
	sessionData, err := store.GetSession(ctx, req.SessionID)
	if err != nil {
		result = "request unauthorized"
		reason = "session_store_error"
		logErr = err
		unauthorizedResponse(ctx, w, req)
		return
	}

	// parse django session to extract user info
	userInfo, err := session.ParseDjangoSession(ctx, []byte(sessionData))
	if err != nil {
		result = "request unauthorized"
		reason = "session_parse_error"
		logErr = err
		unauthorizedResponse(ctx, w, req)
		return
	}

	result = "request authorized"
	userID = userInfo.UserID
	sessionID = req.SessionID

	// return success response with user id
	w.WriteHeader(http.StatusOK)
}
