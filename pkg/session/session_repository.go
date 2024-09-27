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
	_, err = tx.Exec(`DELETE FROM luser_session`)
	if err != nil {
		return err
	}
	for _, session := range SessionsMap {
		insertSession(session, tx)
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func SaveSessionInDb(sesh *Session) error {
	conn, err := infra.CreateConnection()
	if err != nil {
		return err
	}
	SessionsMap[sesh.SessionId] = sesh
	defer conn.Close()
	return insertSession(sesh, conn)
}

func RestoreSessionsFromDb() error {
	sql, err := infra.CreateConnection()
	if err != nil {
		return err
	}
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
		err := rows.Scan(&session.SessionId, &session.User.Id, &session.LastUsed, &session.CreatedAt, &session.User.Username)
		if err != nil {
			return err
		}
		SessionsMap[session.SessionId] = session
	}
	return nil
}

func insertSession(session *Session, quueryable infra.Queryable) error {
	_, err := quueryable.Exec(`INSERT INTO luser_session (sessionId, luserId, lastUsed) VALUES (?, ?, ?)`, session.SessionId, session.User.Id, session.LastUsed)
	if err != nil {
		return err
	}
	return nil
}

func deleteSessionById(sessionId string, quueryable infra.Queryable) error {
	_, err := quueryable.Exec(`DELETE FROM luser_session WHERE sessionId = ?`, sessionId)
	if err != nil {
		return err
	}
	return nil
}
