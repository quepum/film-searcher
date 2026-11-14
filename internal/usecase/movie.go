package usecase

import (
	"context"
	"errors"
	"film-searcher/internal/domain"
	"fmt"
)

type Service struct {
	provider MovieProvider
	repo     MovieRepository
}

func New(provider MovieProvider, repo MovieRepository) *Service {
	return &Service{
		provider: provider,
		repo:     repo,
	}
}

func (s *Service) GetLastMovie(ctx context.Context) (domain.Movie, error) {
	movie, err := s.repo.GetLast(ctx)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMovieNotFound):
			return domain.Movie{}, domain.ErrMovieNotFound
		default:
			return domain.Movie{}, err
		}
	}

	return movie, nil
}

func (s *Service) LoadMovies(ctx context.Context, from, to int) ([]domain.Movie, error) {
	if from <= 0 || to <= 0 || from > to {
		return []domain.Movie{}, fmt.Errorf("%w: from=%d to=%d", domain.ErrInvalidRange, from, to)
	}

	return s.repo.GetRange(ctx, from, to)
}

func (s *Service) findLastOnStand(ctx context.Context) (int, error) {
	left := 1
	right := 40000
	lastFound := -1

	for left <= right {
		mid := (left + right) / 2
		_, err := s.provider.GetMovie(ctx, mid)
		if err != nil {
			if errors.Is(err, domain.ErrMovieNotFound) {
				right = mid - 1
				continue
			}
			return 0, err
		}
		lastFound = mid
		left = mid + 1
	}

	if lastFound < 1 {
		return -1, domain.ErrMovieNotFound
	}
	return lastFound, nil
}

func (s *Service) Sync(ctx context.Context) error {
	last, err := s.findLastOnStand(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrMovieNotFound) {
			return nil
		}
		return err
	}

	existing, err := s.repo.IDs(ctx)
	if err != nil {
		return fmt.Errorf("load existing ids: %w", err)
	}
	have := make(map[int]bool, len(existing))
	for _, id := range existing {
		have[id] = true
	}

	for id := 1; id <= last; id++ {
		if have[id] {
			continue
		}
		m, err := s.provider.GetMovie(ctx, id)
		if err != nil {
			if errors.Is(err, domain.ErrMovieNotFound) {
				continue
			}
			return fmt.Errorf("fetch movie %d: %w", id, err)
		}
		if err := s.repo.Save(ctx, m); err != nil {
			return fmt.Errorf("save movie %d: %w", id, err)
		}
	}
	return nil
}
