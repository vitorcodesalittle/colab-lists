package user

type User struct {
	Id           int64
	Username     string
	PasswordHash string
	PasswordSalt string
	Online       bool
}
