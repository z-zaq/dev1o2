package models

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
	Role     string // "user" or "admin"
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}
