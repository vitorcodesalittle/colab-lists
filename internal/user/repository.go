package user

type UsersRepository interface {
	GetAll() ([]User, error)
	Search(query string) ([]*User, error)
	Get(userId int64) (User, error)
	CreateUser(username, password, email string) (User, error)
	UnsafeGetByUsername(username string) (*User, error)
	ComparePassword(password []byte, hashedPasswowrd []byte) bool
	FindByEmail(email string) (User, error)
}
