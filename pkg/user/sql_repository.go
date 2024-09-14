package user

import (
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
	stmt, err := conn.Prepare(`INSERT INTO luser (username, passwordHash, passwordSalt) VALUES (?, ?, ?) RETURNING luserId`)
	defer stmt.Close()
	if err != nil {
		return User{}, err
	}
	passwordHash, err := hashPassword([]byte(password))
	if err != nil {
		return User{}, err
	}
	rs, err := stmt.Exec(username, passwordHash, "")
	if err != nil {
		return User{}, err
	}
	userId, err := rs.LastInsertId()
	if err != nil {
		return User{}, err
	}
	if err != nil {
		return User{}, err
	}
	conn.Close()
	return s.Get(userId)
}

func (s *SqlUsersRepository) Get(id int64) (User, error) {
	conn, err := infra.CreateConnection()
	defer conn.Close()
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
	return ScanUser(row)
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
func (s *SqlUsersRepository) GetByUsername(username string) (User, error) {
	conn, err := infra.CreateConnection()
	defer conn.Close()
	if err != nil {
		return User{}, err
	}
	stmt, err := conn.Prepare(`SELECT * FROM luser WHERE username = ?`)
	defer stmt.Close()
	if err != nil {
		return User{}, err
	}
	row := stmt.QueryRow(username)
	if row.Err() != nil {
		return User{}, row.Err()
	}
	return ScanUser(row)
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
