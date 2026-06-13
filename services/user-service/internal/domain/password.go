package domain

import (
	"database/sql/driver"
	"fmt"
)

type PasswordHasher interface {
	Hash(password string) ([]byte, error)
	Compare(password string, hash []byte) error
}

type Password struct {
	Text *string
	Hash []byte
}

func (p *Password) GenerateHash(text string, hasher PasswordHasher) error {
	hash, err := hasher.Hash(text)
	if err != nil {
		return err
	}

	p.Hash = hash
	p.Text = &text

	return nil
}

func (p *Password) IsCorrect(plainPassword string, hasher PasswordHasher) bool {
	err := hasher.Compare(plainPassword, p.Hash)
	return err == nil
}

func (p Password) Value() (driver.Value, error) {
	if len(p.Hash) == 0 {
		return nil, nil
	}
	return string(p.Hash), nil
}

func (p *Password) Scan(value any) error {
	if value == nil {
		p.Hash = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		p.Hash = []byte(v)
		return nil
	case []byte:
		p.Hash = v
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into Password", value)
	}
}
