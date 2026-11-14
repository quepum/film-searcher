-- name: ListMoviesRange :many
SELECT id, title, year, director
FROM movies
WHERE id BETWEEN $1 AND $2
ORDER BY id;

-- name: GetLastMovie :one
SELECT id, title, year, director
FROM movies
ORDER BY id DESC
    LIMIT 1;

-- name: UpsertMovie :exec
INSERT INTO movies (id, title, year, director)
VALUES ($1, $2, $3, $4)
    ON CONFLICT (id) DO UPDATE
                            SET title    = EXCLUDED.title,
                            year     = EXCLUDED.year,
                            director = EXCLUDED.director;

-- name: ListMovieIDs :many
SELECT id FROM movies ORDER BY id;