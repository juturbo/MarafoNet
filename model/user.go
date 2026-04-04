package model

import "golang.org/x/crypto/bcrypt"

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func NewUser(name string, password string) (*User, error) {
	return &User{Name: name, Password: password}, nil
}

func (u *User) GeneratePasswordHash() (passwordHash string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	return string(hash), err
}

func (u *User) CheckPassword(hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(u.Password)) == nil
}
