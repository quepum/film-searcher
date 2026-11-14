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
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Movie{}, domain.ErrMovieNotFound
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
	for _, id := range ids {
		result = append(result, int(id))
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

func toDomain(m db.Movie) domain.Movie {
	return domain.Movie{
		ID:       int(m.ID),
		Title:    m.Title,
		Year:     int(m.Year),
		Director: m.Director,
	}
}
