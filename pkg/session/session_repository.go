package session

import (

	"vilmasoftware.com/colablists/pkg/infra"
	"vilmasoftware.com/colablists/pkg/user"
)

func SaveSessionsInDb() error {
	sql, err := infra.CreateConnection()
	if err != nil {
		return err
	}
	defer sql.Close()
	tx, err := sql.Begin()
	if err != nil {
		return err
	}
    defer tx.Rollback()
    ///println("Deleting all sessions")
    ///_, err = tx.Exec(`DELETE FROM luser_session`)
	///if err != nil {
	///	return infra.ErrorRollback(err, tx)
	///}
	for _, session := range SessionsMap {
		_, err := tx.Exec(`INSERT INTO luser_session (sessionId, luserId, lastUsed) VALUES (?, ?, ?)`, session.SessionId, session.User.Id, session.LastUsed)
		if err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func RestoreSessionsFromDb() error {
	sql, err := infra.CreateConnection()
	defer sql.Close()
	if err != nil {
		return err
	}

	rows, err := sql.Query("SELECT session.*, lu.username FROM luser_session session LEFT JOIN luser lu ON session.luserId = lu.luserId")
    if err != nil {
        return err
    }
    defer rows.Close()
	for rows.Next() {
        user := user.User{}
		session := &Session{User: &user}
		err := rows.Scan(&session.SessionId, &session.User.Id, &session.LastUsed, &session.User.Username)
		if err != nil {
			return err
		}
		SessionsMap[session.SessionId] = session
	}

	return nil
}
