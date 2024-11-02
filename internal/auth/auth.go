package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn).UTC()),
		Subject: userID.String(),
	})

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

const (
	ErrInvalidToken = "invalid token"
	ErrExpiredToken = "expired token"
	ErrSignatureInvalid = "signature is invalid"
)

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.UUID{}, fmt.Errorf("%s: %w", ErrExpiredToken, err)
		}
		return uuid.UUID{}, fmt.Errorf("%s: %w", ErrInvalidToken, err)
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", ErrSignatureInvalid, err)
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", ErrSignatureInvalid, err)
	}

	return userID, nil
}

const (
	ErrInvalidBearerToken = "invalid bearer token"
	ErrEmptyBearerToken = "empty bearer token"
	ErrNoAuthorizationHeader = "no authorization header"
)

func GetBearerToken(headers http.Header) (string, error) {
	authorizationHeader := headers.Get("Authorization")
	if authorizationHeader == "" {
		return "", errors.New(ErrNoAuthorizationHeader)
	}

	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return "", errors.New(ErrInvalidBearerToken)
	}
	tokenStringSlice := strings.SplitN(authorizationHeader, "Bearer ", 2)
	tokenString := tokenStringSlice[1]
	if tokenString == "" {
		return "", errors.New(ErrEmptyBearerToken)
	}

	return tokenStringSlice[1], nil
}