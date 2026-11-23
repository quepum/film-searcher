package postgres

import (
	"context"
	"errors"
	"film-searcher/internal/db"
	"film-searcher/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieRepository struct {
	q *db.Queries
}

func New(pool *pgxpool.Pool) *MovieRepository {
	return &MovieRepository{
		q: db.New(pool),
	}
}

func (r *MovieRepository) GetRange(ctx context.Context, from, to int) ([]domain.Movie, error) {
	movies, err := r.q.ListMoviesRange(ctx, db.ListMoviesRangeParams{
		ID:   int64(from),
		ID_2: int64(to),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domain.Movie, len(movies))
	for i, m := range movies {
		result[i] = toDomain(m)
	}
	return result, nil

}

func (r *MovieRepository) GetLast(ctx context.Context) (domain.Movie, error) {
	movie, err := r.q.GetLastMovie(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Movie{}, domain.ErrMovieNotFound
		}
		return domain.Movie{}, err
	}
	result := toDomain(movie)
	return result, nil
}

func (r *MovieRepository) IDs(ctx context.Context) ([]int, error) {
	ids, err := r.q.ListMovieIDs(ctx)
	if err != nil {
		return []int{}, err
	}
	result := make([]int, len(ids))
	for i, id := range ids {
		result[i] = int(id)
	}

	return result, nil
}

func (r *MovieRepository) Save(ctx context.Context, movie domain.Movie) error {
	return r.q.UpsertMovie(ctx, db.UpsertMovieParams{
		ID:       int64(movie.ID),
		Title:    movie.Title,
		Year:     int32(movie.Year),
		Director: movie.Director,
	})
}

func (r *MovieRepository) All(ctx context.Context) ([]domain.Movie, error) {
	movies, err := r.q.ListAllMovies(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Movie, len(movies))
	for i, m := range movies {
		result[i] = toDomain(m)
	}
	return result, nil
}

func (r *MovieRepository) GetByIDs(ctx context.Context, ids []int) ([]domain.Movie, error) {
	if len(ids) == 0 {
		return []domain.Movie{}, nil
	}

	ids64 := make([]int64, len(ids))
	for i, id := range ids {
		ids64[i] = int64(id)
	}

	movies, err := r.q.ListMoviesByIDs(ctx, ids64)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Movie, len(movies))
	for i, m := range movies {
		result[i] = toDomain(m)
	}
	return result, nil
}

func toDomain(m db.Movie) domain.Movie {
	return domain.Movie{
		ID:       int(m.ID),
		Title:    m.Title,
		Year:     int(m.Year),
		Director: m.Director,
	}
}
