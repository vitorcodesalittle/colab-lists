package user
type UsersRepository interface {
	GetAll() ([]User, error)
	Get(userId int64) (User, error)
  CreateUser(username, password string) (User, error)
  GetByUsername(username string) (User, error)
  ComparePassword(password []byte, hashedPasswowrd []byte) (bool)
}
