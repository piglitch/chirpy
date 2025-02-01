package auth

import (
	"log"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	password_byte := []byte(password)
	hashedPass, err := bcrypt.GenerateFromPassword(password_byte, bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash the password: %s", err)
		return string(hashedPass), err
	}
	return string(hashedPass), nil
}

func CheckPasswordHash(password, hash string) error {
	// hash_byte, err := HashPassword(password) 
	// if err != nil {
	// 	return err
	// }
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}