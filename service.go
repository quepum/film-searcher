package main

import (
	"context"
	"errors"
)

type MovieService struct {
	client *Client
}

func NewMovieService(c *Client) *MovieService {
	return &MovieService{client: c}
}

func (s *MovieService) LoadMovies(ctx context.Context, from, to int) ([]Movie, map[int]error) {
	movies := make([]Movie, 0)
	failed := make(map[int]error)
	for i := from; i <= to; i++ {
		m, err := s.client.FetchMovie(ctx, i)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				failed[i] = err
				continue
			}
			failed[i] = err
			continue
		}
		movies = append(movies, *m)
	}
	return movies, failed
}

func (s *MovieService) GetLastMovie(ctx context.Context) (*Movie, error) {
	return s.client.FindLastMovie(ctx)
}
