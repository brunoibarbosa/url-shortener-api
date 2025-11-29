package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

type UserPasswordEncrypter struct {
	cost int
}

func NewUserPasswordEncrypter(cost int) *UserPasswordEncrypter {
	return &UserPasswordEncrypter{
		cost: cost,
	}
}

func (e *UserPasswordEncrypter) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), e.cost)
	return string(bytes), err
}

func (e *UserPasswordEncrypter) CheckPassword(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
