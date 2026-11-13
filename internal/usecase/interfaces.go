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
