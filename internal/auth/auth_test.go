package auth

import (
	"testing"
	"time"
	"fmt"
	"github.com/google/uuid"
)

func TestAuthenticate(t *testing.T) {
	hash, err := HashPassword("password")
	if err != nil {
		t.Errorf("HashPassword error: %v", err)
	}
	want := true

	got, err := CheckPasswordHash("password", hash)
	if err != nil {
		t.Errorf("CheckPasswordHash error: %v", err)
	}

	if got != want {
		t.Errorf("got %t want %t", got, want)
	}
}

func TestMakeJWT(t *testing.T) {
	// Test the MakeJWT function
	userID := uuid.New()
	tokenSecret := "mysecret"
	expiresIn := 1 * time.Hour

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("MakeJWT error: %v", err)
	}
	if token == "" {
		t.Errorf("MakeJWT returned empty token")
	}
	fmt.Printf("Generated JWT: %s user ID: %s\n", token, userID.String())

	// validate the token
	valid, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Errorf("ValidateJWT error: %v", err)
	}
	if valid == uuid.Nil {
		t.Errorf("ValidateJWT returned nil UUID")
	}
	fmt.Printf("Validated JWT user ID: %s\n", valid.String())
}


func TestBearerToken(t *testing.T) {
	// Test the BearerToken function
	token := "mytoken"
	authHeader := fmt.Sprintf("Bearer %s", token)
	got, err := BearerToken(authHeader)
	if err != nil {
		t.Errorf("BearerToken error: %v", err)
	}
	if got != token {
		t.Errorf("got %s want %s", got, token)
	}
}