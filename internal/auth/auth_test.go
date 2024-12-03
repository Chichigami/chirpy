package auth

import (
	"net/http"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "Hello World"
	result, _ := HashPassword(password)
	if err := CheckPasswordHash(result, password); err != nil {
		t.Fatalf("password and hash does not match")
	}
}

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer 123")
	actual, _ := GetBearerToken(headers)
	expected := "123"
	if actual != expected {
		t.Fatalf("Token String are not equal")
	}
}

func TestMakeRefreshToken(t *testing.T) {
	token, _ := MakeRefreshToken()
	expectedLength := 64
	if len(token) != expectedLength {
		t.Fatalf("token: %v is not %v long, it's %v", token, expectedLength, len(token))
	}
}
