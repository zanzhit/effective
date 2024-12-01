package storage

import "errors"

var (
	ErrSongExists   = errors.New("exists")
	ErrSongNotFound = errors.New("song not found")
)
