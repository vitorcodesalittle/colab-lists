package user

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"vilmasoftware.com/colablists/pkg/infra"
)

type SqlUsersRepository struct{}

// CreateUser implements UsersRepository.
func (s *SqlUsersRepository) CreateUser(username string, password string) (User, error) {
	conn, err := infra.CreateConnection()
	if err != nil {
		return User{}, err
	}
	defer conn.Close()

    if conn.QueryRow(`SELECT * FROM luser WHERE username = ?`, username).Err() != sql.ErrNoRows {
        return User{}, fmt.Errorf("Username \"%s\" already registered. Please choose another one.", username)
    }


	passwordHash, err := hashPassword([]byte(password))
	if err != nil {
		return User{}, err
	}
	rs, err := conn.Exec(`INSERT INTO luser (username, passwordHash, passwordSalt) VALUES (?, ?, ?)`, username, passwordHash, "")
	if err != nil {
		return User{}, err
	}
	userId, err := rs.LastInsertId()
	if err != nil {
		return User{}, err
	}
	user, err := s.GetWithConnection(conn, userId)
	return user, err
}

func (s *SqlUsersRepository) GetWithConnection(conn *sql.DB, id int64) (User, error) {
	stmt, err := conn.Prepare(`SELECT * FROM luser WHERE luserId = ?`)
	defer stmt.Close()
	if err != nil {
		panic(err)
	}
	row := stmt.QueryRow(id)
	if row.Err() != nil {
		return User{}, row.Err()
	}
	result, err := ScanUser(row)
	return result, err
}

func (s *SqlUsersRepository) Get(id int64) (User, error) {
	conn, err := infra.CreateConnection()
	if err != nil {
		return User{}, err
	}
	stmt, err := conn.Prepare(`SELECT * FROM luser WHERE luserId = ?`)
	defer stmt.Close()
	if err != nil {
		panic(err)
	}
	row := stmt.QueryRow(id)
	if row.Err() != nil {
		return User{}, row.Err()
	}
	result, err := ScanUser(row)

	conn.Close()
	return result, err
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

// GetAll implements UsersRepository.
func (s *SqlUsersRepository) GetAll() ([]User, error) {
	conn, err := infra.CreateConnection()
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	rs, err := conn.Query(`SELECT * FROM luser`)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	users := make([]User, 0)
	for rs.Next() {
		user, err := ScanUser(rs)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// GetByUsername implements UsersRepository.
func (s *SqlUsersRepository) GetByUsername(username string) (*User, error) {
	conn, err := infra.CreateConnection()
	defer conn.Close()
	if err != nil {
		return nil, err
	}
	stmt, err := conn.Prepare(`SELECT * FROM luser WHERE username = ?`)
	defer stmt.Close()
	if err != nil {
		return nil, err
	}
	row := stmt.QueryRow(username)
	if row.Err() != nil {
		return nil, row.Err()
	}
    user, err := ScanUser(row)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func hashPassword(password []byte) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return hashedPassword, nil
}

func (s *SqlUsersRepository) ComparePassword(password []byte, hashedPasswowrd []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedPasswowrd, password)
	return err == nil
}
