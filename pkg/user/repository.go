package user
type UsersRepository interface {
	GetAll() ([]User, error)
}

type UsersInMemoryRepository struct {
  users []User
}

func NewUsersInMemoryRepository(users []User) *UsersInMemoryRepository {
  return &UsersInMemoryRepository{users: users}
}

func (r *UsersInMemoryRepository) GetAll() ([]User, error) {
  return r.users, nil
}
