package model

import "golang.org/x/crypto/bcrypt"

type User struct {
	Name     string `json:"name"`
	Password string `json:"-"`
}

func NewUser(name string, password string) (*User, error) {
	return &User{Name: name, Password: password}, nil
}

func (u *User) GeneratePasswordHash() (passwordHash []byte, err error) {
	return bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
}

func (u *User) CheckPassword(hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(u.Password)) == nil
}
