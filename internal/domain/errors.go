package domain

import "errors"

var (
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrEmptyMbox          = errors.New("mbox file is empty or contains no readable emails")
)
