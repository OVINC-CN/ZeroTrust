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
			"traceID":             req.TraceID,
			"traceIdDisplayStyle": "block",
		}
		if req.TraceID == "" {
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
	// log incoming verify request
	logrus.WithContext(ctx).WithFields(
		logrus.Fields{
			"client_ip":  req.ClientIP,
			"method":     req.Method,
			"protocol":   req.Protocol,
			"host":       req.Host,
			"path":       req.Path,
			"trace_id":   req.TraceID,
			"user_agent": req.UserAgent,
			"referer":    req.Referer,
		},
	).Infof("verifying request")

	// check if session id is provided
	if req.SessionID == "" {
		unauthorizedResponse(ctx, w, req)
		return
	}

	// get session data from redis
	sessionData, err := store.GetSession(ctx, req.SessionID)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Warn("failed to get session from store")
		unauthorizedResponse(ctx, w, req)
		return
	}

	// parse django session to extract user info
	userInfo, err := session.ParseDjangoSession(ctx, []byte(sessionData))
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Warn("failed to parse session data")
		unauthorizedResponse(ctx, w, req)
		return
	}

	// log successful authorization
	logrus.WithContext(ctx).WithFields(logrus.Fields{
		"user_id":    userInfo.UserID,
		"session_id": maskSessionID(req.SessionID),
	}).Info("request authorized")

	// return success response with user id
	w.WriteHeader(http.StatusOK)
}
