package user

func UnsafeScanUser(row Scanner) (User, error) {
	user := &User{}
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.PasswordSalt, &user.Email, &user.AvatarUrl)
	if err != nil {
		return User{}, err
	}
	return *user, nil
}

func ScanUser(row Scanner, user *User) error {
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.PasswordSalt, &user.Email, &user.AvatarUrl)
	if err != nil {
		return err
	}
	user.PasswordHash = ""
	return nil
}
