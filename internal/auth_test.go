package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTValidation(t *testing.T){
	userId := uuid.New()
	tokenSecret := "mytoken"
	expiresIn := 5 * time.Second
	tokenString, err := MakeJWT(userId, tokenSecret, expiresIn)
	
	if err != nil {
		t.Fatalf("Error generating a token: %s", err)		
	}
	tokenId, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Log("token invalid")
		return
	}
	t.Log(tokenId)
}