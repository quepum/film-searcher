package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"film-searcher/internal/domain"
)

const searchCacheTTL = time.Hour

func (s *Service) BuildIndex(ctx context.Context) error {
	movies, err := s.repo.All(ctx)
	if err != nil {
		return fmt.Errorf("load movies: %w", err)
	}
	if err := s.index.Build(ctx, movies); err != nil {
		return fmt.Errorf("build search index: %w", err)
	}
	if err := s.cache.Clear(ctx); err != nil {
		return fmt.Errorf("clear search cache: %w", err)
	}
	return nil
}

func (s *Service) SearchMovies(ctx context.Context, q string) ([]domain.Movie, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, domain.ErrInvalidQuery
	}

	key := strings.ToLower(strings.Join(strings.Fields(q), " "))
	if movies, ok, err := s.cache.Get(ctx, key); err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	} else if ok {
		return movies, nil
	}

	words := Tokenize(q)
	if len(words) == 0 {
		return []domain.Movie{}, nil
	}

	ids, err := s.index.SearchIDs(ctx, words)
	if err != nil {
		return nil, fmt.Errorf("search index: %w", err)
	}

	movies, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load movies by ids: %w", err)
	}
	if movies == nil {
		movies = []domain.Movie{}
	}

	_ = s.cache.Set(ctx, key, movies, searchCacheTTL)

	return movies, nil
}
