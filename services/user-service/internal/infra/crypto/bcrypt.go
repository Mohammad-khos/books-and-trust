package crypto

import "golang.org/x/crypto/bcrypt"

type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

func (hasher *BcryptHasher) Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (b *BcryptHasher) Compare( password string , hashedPassword []byte) error {
    return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
}