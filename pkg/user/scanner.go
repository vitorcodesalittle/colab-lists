package user

func ScanUser(row Scanner) (User, error) {
	user := &User{}
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.PasswordSalt, &user.Email, &user.AvatarUrl)
	if err != nil {
		return User{}, err
	}
	return *user, nil
}
