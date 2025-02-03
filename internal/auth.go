package auth

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject: userId.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return ss, err
	}
	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	type MyCustomClaims struct{
		Id uuid.UUID `json:"id"`
		jwt.RegisteredClaims
	}
	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token)(interface{}, error){
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Printf("Error retrieving the token: %s", err)
		return uuid.Nil, err
	}
	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims.Id, nil
	} 
	return uuid.Nil, nil
}

