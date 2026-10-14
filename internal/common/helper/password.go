package helper

import "golang.org/x/crypto/bcrypt"

type IPasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type PasswordHelper struct {
	cost int
}

func NewPasswordHelper() *PasswordHelper {
	return &PasswordHelper{cost: bcrypt.DefaultCost}
}

func (h *PasswordHelper) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *PasswordHelper) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
