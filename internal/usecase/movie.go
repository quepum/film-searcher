package usecase

import (
	"context"
	"errors"
	"film-searcher/internal/domain"
	"fmt"
)

type Service struct {
	provider MovieProvider
}

func New(provider MovieProvider) *Service {
	return &Service{
		provider: provider,
	}
}

func (s *Service) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	left := 1
	right := 100000
	lastFound := -1

	for left <= right {
		mid := (left + right) / 2
		_, err := s.provider.GetMovie(ctx, mid)
		if err != nil {
			if errors.Is(err, domain.ErrMovieNotFound) {
				right = mid - 1
				continue
			}
			return domain.Movie{}, err
		}
		lastFound = mid
		left = mid + 1
	}

	if lastFound < 1 {
		return domain.Movie{}, domain.ErrMovieNotFound
	}

	return s.provider.GetMovie(ctx, lastFound)
}

func (s *Service) LoadMovies(ctx context.Context, from, to int) ([]domain.Movie, error) {
	if from <= 0 || to <= 0 || from > to {
		return []domain.Movie{}, fmt.Errorf("%w: from=%d to=%d", domain.ErrInvalidRange, from, to)
	}

	var movies []domain.Movie
	var failedMovies []domain.MovieError
	for i := from; i <= to; i++ {
		if ctx.Err() != nil {
			return movies, ctx.Err()
		}
		movie, err := s.provider.GetMovie(ctx, i)
		if err != nil {
			failedMovies = append(failedMovies, domain.MovieError{
				ID:  i,
				Err: err})
			continue
		}
		movies = append(movies, movie)
	}

	if len(failedMovies) > 0 {
		return movies, &domain.PartialError{Failed: failedMovies}
	}
	return movies, nil
}
