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
