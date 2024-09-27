package recovery

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"vilmasoftware.com/colablists/pkg/infra"
	"vilmasoftware.com/colablists/pkg/user"
)

type Recovery struct {
	UserRepository user.UsersRepository
}

// TODO: improve this function
func IsPasswordValid(password string) bool {
	return len(password) < 5
}

func (r *Recovery) CreatePasswordRecoveryRequest(email string) error {
	user, err := r.UserRepository.FindByEmail(email)
	if err != nil {
		return err
	}
	if user.Id == 0 {
		return nil
	}
	tokenBytes := make([]byte, 256)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	token := fmt.Sprintf("%x", sha256.New().Sum(tokenBytes))
	db, err := infra.CreateConnection()
	if err != nil {
		return err
	}
	defer db.Close()
	if _, err := db.Exec(`INSERT INTO password_recovery_requests
    (luserId, email, token)
    VALUES (?, ?, ?)`, user.Id, email, token); err != nil {
		return err
	}
	fmt.Printf("Recovery URL is at /password-recovery?token=%s\n", token)
	// TODO: send email

	return nil
}

type passwordRecoveryEntry struct {
	id         int64
	luserId    int64
	issuedAt   time.Time
	consumedAt *time.Time
}

func (r *Recovery) RecoverPassword(password, token string) error {
	db, err := infra.CreateConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	entry := &passwordRecoveryEntry{}
	dbTokenRs := db.QueryRow(`SELECT id, luserId, issuedAt, consumedAt FROM password_recovery_requests where token = ?`, token)

	if err == sql.ErrNoRows {
		return errors.New("invalid token")
	} else if dbTokenRs.Err() != nil {
		return dbTokenRs.Err()
	}
	dbTokenRs.Scan(&entry.id, &entry.luserId, &entry.issuedAt, entry.consumedAt)

	if entry.consumedAt != nil {
		return errors.New("entry already consumed")
	}

	if entry.issuedAt.Add(10 * time.Second).After(time.Now()) {
		return errors.New("entry issued 10 seconds after issuedAt " + entry.issuedAt.String())
	}

	passwordHash, err := user.HashPassword([]byte(password))
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = db.Exec(`UPDATE luser SET passwordHash = ? WHERE luserId = ?`, passwordHash, entry.luserId)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE password_recovery_requests SET consumedAt = ? WHERE id = ?`, time.Now(), entry.id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
