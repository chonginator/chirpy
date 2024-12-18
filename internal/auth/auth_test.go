package auth

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

	validToken, err := MakeJWT(userID, tokenSecret, time.Hour)
	if err != nil {
		t.Fatalf("Couldn't make JWT: %v", err)
	}

	tests := []struct{
		name string
		tokenString string
		tokenSecret string
		errorContains string
	}{
		{
			name: "Valid JWT",
			tokenString: validToken,
			tokenSecret: tokenSecret,
			errorContains: "",
		},
		{
			name: "Invalid JWT",
			tokenString: "notatoken",
			tokenSecret: tokenSecret,
			errorContains: jwt.ErrTokenMalformed.Error(),
		},
		{
			name: "Wrong secret",
			tokenString: validToken,
			tokenSecret: "secret",
			errorContains: jwt.ErrSignatureInvalid.Error(),
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err = ValidateJWT(tc.tokenString, tc.tokenSecret)
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
	tests := []struct{
		name string
		headers http.Header
		expected string
		errorContains string
	}{
		{
			name: "Valid Bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer token_string"},
			},
			expected: "token_string",
			errorContains: "",
		},
		{
			name: "Missing Authorization header",
			headers: http.Header{},
			expected: "",
			errorContains: ErrNoAuthorizationHeader,
		},
		{
			name: "No space",
			headers: http.Header{
				"Authorization": []string{"Bearertoken_string"},
			},
			expected: "",
			errorContains: ErrInvalidAuthHeader.Error(),
		},
		{
			name: "No prefix",
			headers: http.Header{
				"Authorization": []string{"token_string"},
			},
			expected: "",
			errorContains: ErrInvalidAuthHeader.Error(),
		},
		{
			name: "Invalid prefix",
			headers: http.Header{
				"Authorization": []string{"Bear token_string"},
			},
			expected: "",
			errorContains: ErrInvalidAuthHeader.Error(),
		},
		{
			name: "Multiple prefixes",
			headers: http.Header{
				"Authorization": []string{"Bearer token_string Bearer token_string"},
			},
			expected: "token_string Bearer token_string",
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