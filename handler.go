package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

// @title Film Searcher API
// @version 1.0
// @description HTTP API for films
// @host localhost:8080
// @BasePath /
type Handler struct {
	service *MovieService
}

func NewHandler(s *MovieService) *Handler {
	return &Handler{service: s}
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

// loadMovies godoc
// @Summary Load movies in range
// @Description Load movies from stand in range [from; to]
// @Tags movies
// @Accept json
// @Produce json
// @Param from query int true "from"
// @Param to query int true "to"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string "Bad Request"
// @Router /load-movies [get]
func (h *Handler) loadMovies(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromStr := q.Get("from")
	toStr := q.Get("to")

	from, err := strconv.Atoi(fromStr)
	if err != nil || from <= 0 {
		http.Error(w, "invalid from", http.StatusBadRequest)
		return
	}
	to, err := strconv.Atoi(toStr)
	if err != nil || to <= 0 || from > to {
		http.Error(w, "invalid to", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	type failed struct {
		ID    int    `json:"id"`
		Error string `json:"error"`
	}

	resp := struct {
		Movies []Movie  `json:"movies"`
		Failed []failed `json:"failed"`
	}{
		Movies: []Movie{},
		Failed: []failed{},
	}

	movies, failedMap := h.service.LoadMovies(ctx, from, to)
	resp.Movies = movies
	for id, ferr := range failedMap {
		if errors.Is(ferr, ErrNotFound) {
			resp.Failed = append(resp.Failed, failed{ID: id, Error: "not found"})
			continue
		}
		resp.Failed = append(resp.Failed, failed{ID: id, Error: ferr.Error()})
	}
	writeJSON(w, http.StatusOK, resp)
}

// lastMovie godoc
// @Summary Get last movie
// @Description Find last available movie on stand
// @Tags movies
// @Accept json
// @Produce json
// @Success 200 {object} Movie
// @Failure 502 {string} string "Bad Gateway"
// @Failure 504 {string} string "Gateway Timeout"
// @Router /last-movie [get]
func (h *Handler) lastMovie(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	movie, err := h.service.GetLastMovie(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "timeout", http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "no movies found", http.StatusBadGateway)
			return
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, movie)
}
