package domain

import (
	"errors"
	"fmt"
)

var (
	ErrMovieNotFound = errors.New("movie not found")
	ErrInvalidRange  = errors.New("invalid range")
	ErrInvalidQuery  = errors.New("invalid search query")
)

type MovieError struct {
	ID  int
	Err error
}

type PartialError struct {
	Failed []MovieError
}

func (e *PartialError) Error() string {
	return fmt.Sprintf("%d movie(s) failed to load", len(e.Failed))
}
