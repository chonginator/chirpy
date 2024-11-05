package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string
const (
	TokenTypeAccess TokenType = "chirpy-access"
)

var ErrNoAuthHeader = errors.New("no authorization header included in request")
var ErrInvalidAuthHeader = errors.New("malformed authorization header")

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

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	signingKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: string(TokenTypeAccess),
		IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour).UTC()),
		Subject: userID.String(),
	})

	return token.SignedString(signingKey)
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return userID, nil
}

const (
	ErrInvalidBearerToken = "invalid bearer token"
	ErrEmptyBearerToken = "empty bearer token"
	ErrNoAuthorizationHeader = "no authorization header"
)

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeader
	}

	splitAuthHeader := strings.SplitN(authHeader, " ", 2)
	if len(splitAuthHeader) < 2 || splitAuthHeader[0] != "Bearer" {
		return "", ErrInvalidAuthHeader
	}

	return splitAuthHeader[1], nil

}

func MakeRefreshToken() (string, error) {
	token := make([]byte, 256)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	encodedToken := hex.EncodeToString(token)
	return encodedToken, nil
}