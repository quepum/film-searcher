package http

import (
	"context"
	"encoding/json"
	"errors"
	"film-searcher/internal/domain"
	"film-searcher/internal/usecase"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Handler
// @title Film Searcher API
// @version 1.0
// @description HTTP API for films
// @host localhost:8080
// @BasePath /
type Handler struct {
	uc      usecase.MovieUseCase
	timeout time.Duration
}

func New(uc usecase.MovieUseCase, timeout time.Duration) *Handler {
	return &Handler{
		uc:      uc,
		timeout: timeout,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /load-movies", h.loadMovies)
	mux.HandleFunc("GET /last-movie", h.lastMovie)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// ignore
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRange):
		http.Error(w, "invalid range", http.StatusBadRequest)
	case errors.Is(err, context.DeadlineExceeded):
		http.Error(w, "timeout", http.StatusGatewayTimeout)
	case errors.Is(err, domain.ErrMovieNotFound):
		http.Error(w, "no movies found", http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
}

// loadMovies
// @Summary Load movies in range
// @Description Load movies from stand in range [from; to]
// @Tags movies
// @Accept json
// @Produce json
// @Param from query int true "from"
// @Param to query int true "to"
// @Success 200 {array} domain.Movie
// @Failure 400 {string} string "Bad Request"
// @Router /load-movies [get]
func (h *Handler) loadMovies(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")

	from, err := strconv.Atoi(fromStr)
	if err != nil {
		http.Error(w, "invalid from", http.StatusBadRequest)
		return
	}
	to, err := strconv.Atoi(toStr)
	if err != nil {
		http.Error(w, "invalid to", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	movies, err := h.uc.LoadMovies(ctx, from, to)
	var partial *domain.PartialError
	if errors.As(err, &partial) {
		ids := make([]int, 0, len(partial.Failed))
		for el := range partial.Failed {
			ids = append(ids, partial.Failed[el].ID)
		}
		sort.Ints(ids)
		parts := make([]string, 0, len(ids))
		for _, id := range ids {
			parts = append(parts, strconv.Itoa(id))
		}
		w.Header().Set("X-Failed-IDs", strings.Join(parts, ","))
		w.Header().Set("X-Failed-Count", strconv.Itoa(len(ids)))
		err = nil
	}

	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, movies)
}

// lastMovie
// @Summary Get last movie
// @Description Find last available movie on stand
// @Tags movies
// @Accept json
// @Produce json
// @Success 200 {object} domain.Movie
// @Failure 404 {string} string "Not Found"
// @Failure 502 {string} string "Bad Gateway"
// @Failure 504 {string} string "Gateway Timeout"
// @Router /last-movie [get]
func (h *Handler) lastMovie(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	movie, err := h.uc.GetLastMovie(ctx)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, movie)
}
