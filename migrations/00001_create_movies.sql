-- +goose Up
CREATE TABLE movies (
                        id bigint PRIMARY KEY,
                        title text NOT NULL,
                        year integer NOT NULL,
                        director text NOT NULL
);

-- +goose Down
DROP TABLE movies;