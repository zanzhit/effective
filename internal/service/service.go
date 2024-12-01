package service

import "errors"

var (
	ErrInvalidVerseNumber = errors.New("invalid verse number")
	ErrInvalidDateFormat  = errors.New("invalid date format")
	ErrEmptyUpdate        = errors.New("update data is epmty")
)
