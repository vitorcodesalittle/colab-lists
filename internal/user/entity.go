package user

import "fmt"

type User struct {
	Id           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	PasswordSalt string `json:"passwordSalt"`
	Online       bool   `json:"online"`
	Email        string `json:"email"`
	AvatarUrl    string `json:"avatarUrl"`
}

func (user *User) String() string {
	return fmt.Sprintf("User{\nId: %v,\nUsername: %v,\nPasswordHash: %v,\nPasswordSalt: %v,\nOnline: %v,\n}", user.Id, user.Username, user.PasswordHash, user.PasswordSalt, user.Online)
}
