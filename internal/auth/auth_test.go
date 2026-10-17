package auth

import (
	"testing"
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