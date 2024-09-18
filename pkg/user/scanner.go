package user

func ScanUser(row Scanner) (User, error) {
	println("Scanning user")
	user := &User{}
	err := row.Scan(&user.Id, &user.Username, &user.PasswordHash, &user.PasswordSalt)
	if err != nil {
		println("FAILED TO READ A USER ")
		return User{}, err
	}
	println("User ", user.String())
	return *user, nil
}
