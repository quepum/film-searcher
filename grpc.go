package main

import (
	"context"
	"errors"
	pb "film-searcher/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedMovieServiceServer
	service *MovieService
}

func NewServer(s *MovieService) *Server {
	return &Server{service: s}
}

func (s *Server) LoadMovies(ctx context.Context, req *pb.LoadMoviesRequest) (*pb.LoadMoviesResponse, error) {
	from := int(req.From)
	to := int(req.To)
	if from <= 0 || to <= 0 || from > to {
		return nil, status.Error(codes.InvalidArgument, "invalid from/to")
	}

	movies, failed := s.service.LoadMovies(ctx, from, to)

	resp := &pb.LoadMoviesResponse{
		Movies: make([]*pb.Movie, 0, len(movies)),
		Failed: make([]*pb.MovieError, 0, len(failed)),
	}

	for _, m := range movies {
		resp.Movies = append(resp.Movies, &pb.Movie{
			Id:       int32(m.ID),
			Title:    m.Title,
			Year:     int32(m.Year),
			Director: m.Director,
		})
	}
	for id, ferr := range failed {
		msg := ferr.Error()
		if errors.Is(ferr, ErrNotFound) {
			msg = "not found"
		}
		resp.Failed = append(resp.Failed, &pb.MovieError{
			Id:    int32(id),
			Error: msg,
		})
	}

	return resp, nil
}

func (s *Server) GetLastMovie(ctx context.Context, req *pb.GetLastMovieRequest) (*pb.Movie, error) {
	movie, err := s.service.GetLastMovie(ctx)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, status.Error(codes.NotFound, "no movies found")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "timeout")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Movie{
		Id:       int32(movie.ID),
		Title:    movie.Title,
		Year:     int32(movie.Year),
		Director: movie.Director,
	}, nil
}
