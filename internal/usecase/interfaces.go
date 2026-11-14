package usecase

import "context"
import "film-searcher/internal/domain"

type MovieProvider interface {
	GetMovie(ctx context.Context, id int) (domain.Movie, error)
}

type MovieUseCase interface {
	LoadMovies(ctx context.Context, from, to int) ([]domain.Movie, error)
	GetLastMovie(ctx context.Context) (domain.Movie, error)
}

type MovieRepository interface {
	GetRange(ctx context.Context, from, to int) ([]domain.Movie, error)
	GetLast(ctx context.Context) (domain.Movie, error)
	Save(ctx context.Context, movie domain.Movie) error
	IDs(ctx context.Context) ([]int, error)
}
