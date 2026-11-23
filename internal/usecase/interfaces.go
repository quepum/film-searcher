package usecase

import (
	"context"
	"time"

	"film-searcher/internal/domain"
)

type MovieProvider interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from, to int) ([]domain.Movie, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
	SearchMovies(ctx context.Context, q string) ([]domain.Movie, error)
}

type MovieRepository interface {
	GetRange(ctx context.Context, from, to int) ([]domain.Movie, error)
	GetLast(ctx context.Context) (domain.Movie, error)
	Save(ctx context.Context, movie domain.Movie) error
	IDs(ctx context.Context) ([]int, error)
	All(ctx context.Context) ([]domain.Movie, error)
	GetByIDs(ctx context.Context, ids []int) ([]domain.Movie, error)
}

type SearchIndex interface {
	Build(ctx context.Context, movies []domain.Movie) error
	SearchIDs(ctx context.Context, words []string) ([]int, error)
}

type SearchCache interface {
	Get(ctx context.Context, key string) ([]domain.Movie, bool, error)
	Set(ctx context.Context, key string, movies []domain.Movie, ttl time.Duration) error
	Clear(ctx context.Context) error
}
