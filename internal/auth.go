package auth

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
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

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	
	claims := &jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: "",
		ExpiresAt: jwt.NewNumericDate(time.Unix(int64(expiresIn.Seconds()), 0)),
		Subject: userId,
	}
	token := jwt.NewNumericDate(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return ss, err
	}
	return ss, nil
}