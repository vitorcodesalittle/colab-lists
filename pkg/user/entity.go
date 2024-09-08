package user


type User struct {
	Id     int
	Name   string
	Email  string
	Online bool
}

func (u *User) String() string {
  return u.Name
}

