package session

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"vilmasoftware.com/colablists/pkg/config"
	"vilmasoftware.com/colablists/pkg/infra"
	"vilmasoftware.com/colablists/pkg/user"
)

const CleanerInterval time.Duration = 5 * time.Minute

var SessionsMap map[string]*Session = make(map[string]*Session)

type Session struct {
	*user.User
	SessionId string
	LastUsed  time.Time
	CreatedAt time.Time
}

func generateRandomBytes(n int) []byte {
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
	for {
		sessionIdBytes := base64.RawStdEncoding.EncodeToString(generateRandomBytes(128))
		sessionId := string(sessionIdBytes)
		if _, ok := SessionsMap[sessionId]; !ok {
			return sessionId
		}
	}
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

func SessionPeriodicallyCleaner() {
	ticker := time.NewTicker(CleanerInterval)
	for {
		select {
		case <-ticker.C:
			for sessionId, session := range SessionsMap {
				if time.Since(session.LastUsed) > config.GetConfig().SessionTimeout {
					db, err := infra.CreateConnection()
					if err != nil {
						println("Failed to delete session " + sessionId + " because of database connection error " + err.Error())
					}
					err = deleteSessionById(sessionId, db)
					if err != nil {
						println("Failed to delete session! " + sessionId)
					}
					delete(SessionsMap, sessionId)
				}
			}
		}
	}
}
