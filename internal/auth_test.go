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

func TestHeaders(t *testing.T){
	Header := map[string][]string{
    "Authorization": {"Bearer token_string"},
	}
	cleanStr, err := GetBearerToken(Header)
	if err != nil {
		t.Logf("could not clean headers: %s", err)
	}
	t.Log(cleanStr)
}

