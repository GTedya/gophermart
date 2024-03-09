package handlers

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

func getIDFromToken(bearerToken string, key []byte) (int64, error) {
	tokenString := strings.Split(bearerToken, " ")[1]

	token, err := jwt.ParseWithClaims(tokenString, &Token{}, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	user := token.Claims.(*Token)

	return user.UserID, err
}

func tokenCreate(secretKey []byte, login int64) (string, error) {
	var tokenClaim = Token{
		UserID: login,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpires).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaim)

	tok, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("token signing error: %w", err)
	}
	return tok, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
