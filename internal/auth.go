package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
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
		jwt.RegisteredClaims
	}

	token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token)(interface{}, error){
		log.Print("55 token parse with claims", token, tokenSecret)
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Printf("Error retrieving the token: %s", err)
		return uuid.Nil, err
	}
	claims, ok := token.Claims.(*MyCustomClaims)
	if !ok && !token.Valid {
		log.Print(claims, token.Valid, " validate 62")
		return uuid.Nil, nil
	} 
	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		log.Printf("failed to parse: %s", err)
		return uuid.Nil, err
	}
	return id, nil
}

func GetBearerToken(headers http.Header)(string, error){
	if headers == nil {
		return "", errors.New("no headers were sent")
	}
	bearerString := headers.Get("Authorization")
	strippedString := bearerString[6:]
	cleanStringArray := strings.Split(strippedString, " ") 
	cleanedString := strings.Join(cleanStringArray, "")
	log.Print(bearerString, 74)
	log.Print(strippedString, 75)
	log.Print(cleanStringArray, 76)
	log.Print(cleanedString, 77)
	return cleanedString, nil
}

func MakeRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	ecodedStr := hex.EncodeToString(b)
	return ecodedStr, nil
}