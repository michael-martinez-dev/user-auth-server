package util

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrWikiAlreadyExists  = errors.New("wiki page already exists")
	ErrEmptyUser          = errors.New("user can't be empty")
	ErrEmptyName          = errors.New("name can't be empty")
	ErrEmptyPassword      = errors.New("password can't be empty")
	ErrInvalidAuthToken   = errors.New("invalid auth-token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
)
