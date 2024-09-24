package user

import "fmt"

type User struct {
	Id           int64
	Username     string
	PasswordHash string
	PasswordSalt string
	Online       bool
	Email        string
	AvatarUrl    string
}

func (user *User) String() string {
	return fmt.Sprintf("User{\nId: %v,\nUsername: %v,\nPasswordHash: %v,\nPasswordSalt: %v,\nOnline: %v,\n}", user.Id, user.Username, user.PasswordHash, user.PasswordSalt, user.Online)
}
