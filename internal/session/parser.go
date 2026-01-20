package session

import (
	"bytes"
	"errors"

	"github.com/nlpodyssey/gopickle/pickle"
	"github.com/nlpodyssey/gopickle/types"
)

var (
	ErrInvalidSession = errors.New("invalid session data")
	ErrUserNotFound   = errors.New("user not found in session")
)

type UserInfo struct {
	UserID   string
	Backend  string
	UserHash string
}

func ParseDjangoSession(data []byte) (*UserInfo, error) {
	// create reader and unpickler for session data
	reader := bytes.NewReader(data)
	unpickler := pickle.NewUnpickler(reader)

	// unpickle session data
	result, err := unpickler.Load()
	if err != nil {
		return nil, ErrInvalidSession
	}

	// session data should be a dict
	sessionDict, ok := result.(*types.Dict)
	if !ok {
		return nil, ErrInvalidSession
	}

	userInfo := &UserInfo{}

	// extract user id from session (required field)
	if userID, ok := sessionDict.Get("_auth_user_id"); ok {
		userInfo.UserID = toString(userID)
	} else {
		return nil, ErrUserNotFound
	}

	// extract auth backend from session (optional field)
	if backend, ok := sessionDict.Get("_auth_user_backend"); ok {
		userInfo.Backend = toString(backend)
	}

	// extract user hash from session (optional field)
	if userHash, ok := sessionDict.Get("_auth_user_hash"); ok {
		userInfo.UserHash = toString(userHash)
	}

	return userInfo, nil
}

func toString(v interface{}) string {
	// handle different types for value conversion
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return ""
	}
}
