package user

import "errors"

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"username"`
	Password string `json:"-"`
	Created  string `json:"-"`
}

var (
	ErrUserNotFound = errors.New("user doesnt exist")
	ErrBadPass      = errors.New("bad user password")
	ErrUserExist    = errors.New("this login exists")
)
