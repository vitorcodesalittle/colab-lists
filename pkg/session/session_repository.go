package session

import (
	"log"

	"vilmasoftware.com/colablists/pkg/infra"
	"vilmasoftware.com/colablists/pkg/user"
)

func SaveSessionsInDb() error {
	sql, err := infra.CreateConnection()
	if err != nil {
		return err
	}
    tx, err := sql.Begin()
    if err != nil {
        return err
    }
    _, err = tx.Exec(`DELETE FROM luser_session`)
    if err != nil {
        return infra.ErrorRollback(err, tx)
    }
	for _, session := range SessionsMap {
        _, err := tx.Exec(`INSERT INTO luser_session (sessionId, luserId, lastUsed) VALUES (?, ?, ?)`, session.SessionId, session.User.Id, session.LastUsed)
        if err != nil {
            return infra.ErrorRollback(err, tx)
        }
	}
    if err = tx.Commit(); err != nil {
        log.Fatal(err)
    }
	return nil
}

func RestoreSessionsFromDb() error {
	sql, err := infra.CreateConnection()
    defer sql.Close()
	if err != nil {
		return err
	}
    tx, err := sql.Begin()
    if err != nil {
        return err
    }

    rows, err := tx.Query("SELECT session.*, lu.username FROM luser_session session LEFT JOIN luser lu ON session.luserId = lu.luserId")
    if err != nil {
        return err;
    }

    for rows.Next() {
        session := &Session{User: user.User{}}
        err := rows.Scan(&session.SessionId, &session.User.Id, &session.LastUsed, &session.User.Username)
        if err != nil {
            log.Println("Error scanning session")
            return err;
        }
        SessionsMap[session.SessionId] = *session
    }


    return nil;
}

