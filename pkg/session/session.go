package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"vilmasoftware.com/colablists/pkg/user"
)


var SessionsMap map[string]*Session = make(map[string]*Session)

type Session struct {
	*user.User
	SessionId string
	LastUsed  time.Time
}

func GenerateRandomBytes(n int) []byte {
	if n == 0 {
		panic("n must be greater than 0")
	}
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func GetSessionId() string {
	for true {
		sessionIdBytes := base64.RawStdEncoding.EncodeToString(GenerateRandomBytes(128))
		sessionId := string(sessionIdBytes)
		if _, ok := SessionsMap[sessionId]; !ok {
			return sessionId
		}
	}
	return ""
}

func GetUserFromSession(r *http.Request) (*user.User, error) {
	sessionId, err := r.Cookie("SESSION")
	if err != nil {
		return nil, err
	}
	session, ok := SessionsMap[sessionId.Value]
	if !ok {
		return nil, errors.New("Session not found")
	}
	return session.User, nil
}

