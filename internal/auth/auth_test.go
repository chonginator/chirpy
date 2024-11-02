package auth

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name string
		password string
	}{
		{
			name: "No errors",
			password: "password",
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := HashPassword(tc.password)
			if err != nil {
				t.Errorf("Test %v - '%s': FAIL: error hashing password: %v", i, tc.name, err)
				return
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password1 := "password"
	password2 := "password2"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name string
		password string
		hash string
		wantErr bool
	} {
		{
			name: "Correct password",
			password: password1,
			hash: hash1,
			wantErr: false,
		},
		{
			name: "Incorrect password",
			password: "wrong password",
			hash: hash1,
			wantErr: true,
		},
		{
			name: "Password doesn't match different hash",
			password: password1,
			hash: hash2,
			wantErr: true,
		},
		{
			name: "Empty password",
			password: "",
			hash: hash1,
			wantErr: true,
		},
		{
			name: "Invalid hash",
			password: password1,
			hash: "hashbrown",
			wantErr: true,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckPasswordHash(tc.password, string(tc.hash))
			if (err != nil) != tc.wantErr {
				t.Errorf("Test %v - '%s' FAIL: error = %v, wantErr %v", i, tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "shhhhh"

	expiresIn1, err := time.ParseDuration("-1m")
	if err != nil {
		t.Fatalf("Couldn't parse time duration string: %v", err)
	}
	expiresIn2, err := time.ParseDuration("1m")
	if err != nil {
		t.Fatalf("Couldn't parse time duration string: %v", err)
	}

	jwt1, err := MakeJWT(userID, tokenSecret, expiresIn1)
	if err != nil {
		t.Fatalf("Couldn't make JWT: %v", err)
	}

	jwt2, err := MakeJWT(userID, tokenSecret, expiresIn2)
	if err != nil {
		t.Fatalf("Couldn't make JWT: %v", err)
	}

	tests := []struct{
		name string
		userID uuid.UUID
		tokenSecret string
		expiresIn time.Duration
		jwt string
		errorContains string
	}{
		{
			name: "Expired token",
			userID: userID,
			tokenSecret: tokenSecret,
			expiresIn: expiresIn1,
			jwt: jwt1,
			errorContains: ErrExpiredToken,
		},
		{
			name: "Valid JWT",
			userID: userID,
			tokenSecret: tokenSecret,
			expiresIn: expiresIn2,
			jwt: jwt2,
			errorContains: "",
		},
		{
			name: "Different token secret",
			userID: userID,
			tokenSecret: "secret",
			expiresIn: expiresIn2,
			jwt: jwt2,
			errorContains: ErrInvalidToken,
		},
		{
			name: "Different user ID",
			userID: uuid.New(),
			tokenSecret: "secret",
			expiresIn: expiresIn2,
			jwt: jwt2,
			errorContains: ErrSignatureInvalid,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err = ValidateJWT(tc.jwt, tc.tokenSecret)
			if err != nil && !strings.Contains(err.Error(), tc.errorContains) {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
			} else if err != nil && tc.errorContains == "" {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
			} else if err == nil && tc.errorContains != "" {
				t.Errorf("Test %v - '%s' FAIL: expected error containing: '%v' got none.", i, tc.name, tc.errorContains)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	tokenString := "token_string"
	tokenString2 := fmt.Sprintf("%s Bearer %s", tokenString, tokenString)

	headers1 := http.Header{}
	headers1.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	headers2 := http.Header{}

	headers3 := http.Header{}
	headers3.Add("Authorization", fmt.Sprintf("Bearer%s", tokenString))

	headers4 := http.Header{}
	headers4.Add("Authorization", tokenString)

	headers5 := http.Header{}
	headers5.Add("Authorization", fmt.Sprintf("Bear %s", tokenString))

	headers6 := http.Header{}
	headers6.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString2))

	tests := []struct{
		name string
		headers http.Header
		expected string
		errorContains string
	}{
		{
			name: "Valid Bearer token",
			headers: headers1,
			expected: tokenString,
			errorContains: "",
		},
		{
			name: "No Bearer token",
			headers: headers2,
			expected: "",
			errorContains: ErrNoAuthorizationHeader,
		},
		{
			name: "No space",
			headers: headers3,
			expected: "",
			errorContains: ErrInvalidBearerToken,
		},
		{
			name: "No prefix",
			headers: headers4,
			expected: "",
			errorContains: ErrInvalidBearerToken,
		},
		{
			name: "Invalid prefix",
			headers: headers5,
			expected: "",
			errorContains: ErrInvalidBearerToken,
		},
		{
			name: "Multiple prefixes",
			headers: headers6,
			expected: tokenString2,
			errorContains: "",
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenString, err := GetBearerToken(tc.headers)
			if err != nil && !strings.Contains(err.Error(), tc.errorContains) {
				t.Errorf("Test %v - '%s': FAIL: unexpected error: %v", i, tc.name, err)
				return
			} else if err != nil && tc.errorContains == "" {
				t.Errorf("Test %v - '%s': FAIL: unexpected error: %v", i, tc.name, err)
				return
			} else if err == nil && tc.errorContains != "" {
				t.Errorf("Test %v - '%s' FAIL: expected error containing: '%v' got none.", i, tc.name, tc.errorContains)
				return
			}

			if tokenString != tc.expected {
				t.Errorf("Test %v - '%s' FAIL: expected token string: %s, actual: %s", i, tc.name, tc.expected, tokenString)
				return
			}
		})
	}
}