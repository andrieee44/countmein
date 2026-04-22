package api

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func Hash(plain string) ([]byte, error) {
	var (
		hash []byte
		err  error
	)

	hash, err = bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return nil, err
		}

		return nil, err
	}

	return hash, nil
}

func ComparePlainToHash(plain string, hash []byte) (bool, error) {
	var err error

	err = bcrypt.CompareHashAndPassword(hash, []byte(plain))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
