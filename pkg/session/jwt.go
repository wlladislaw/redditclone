package session

import (
	"errors"
	"redditclone/pkg/hash"
	"redditclone/pkg/user"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWTSessClaims struct {
	*user.User `json:"user"`
	SessionId  string `json:"session"`
	jwt.StandardClaims
}

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrSignToken    = errors.New("cant sign token")
)

func CreateJWT(userID int, login string, sessId string) (string, error) {
	claims := JWTSessClaims{
		User: &user.User{
			ID:    userID,
			Login: login,
		},
		SessionId: sessId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 10 * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(hash.TokenSecret)
	if err != nil {
		return "", ErrSignToken
	}

	return tokenString, nil
}

func CheckJWT(token string) (*JWTSessClaims, error) {
	checkedToken, err := jwt.ParseWithClaims(token, &JWTSessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(hash.TokenSecret), nil
	})
	if err != nil || !checkedToken.Valid {
		return nil, ErrInvalidToken
	}

	claims := checkedToken.Claims.(*JWTSessClaims)
	return claims, nil
}
