package grpc

import (
	"context"
	"errors"
	"film-searcher/internal/domain"
	"film-searcher/internal/usecase"
	pb "film-searcher/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	uc usecase.MovieUseCase
	pb.UnimplementedMovieServiceServer
}

func New(uc usecase.MovieUseCase) *Server {
	return &Server{
		uc: uc,
	}
}

func (s *Server) LoadMovies(ctx context.Context, req *pb.LoadMoviesRequest) (*pb.LoadMoviesResponse, error) {
	from := int(req.From)
	to := int(req.To)

	movies, err := s.uc.LoadMovies(ctx, from, to)
	var partial *domain.PartialError
	if errors.As(err, &partial) {
		err = nil
	}

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRange):
			return nil, status.Error(codes.InvalidArgument, "invalid from/to")
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, "timeout")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	resp := &pb.LoadMoviesResponse{
		Movies: make([]*pb.Movie, 0, len(movies)),
	}
	for _, m := range movies {
		resp.Movies = append(resp.Movies, &pb.Movie{
			Id:       int32(m.ID),
			Title:    m.Title,
			Year:     int32(m.Year),
			Director: m.Director,
		})
	}

	if partial != nil {
		resp.Failed = make([]*pb.MovieError, 0, len(partial.Failed))
		for _, f := range partial.Failed {
			msg := f.Err.Error()
			if errors.Is(f.Err, domain.ErrMovieNotFound) {
				msg = "not found"
			}
			resp.Failed = append(resp.Failed, &pb.MovieError{
				Id:    int32(f.ID),
				Error: msg,
			})
		}
	}

	return resp, nil
}

func (s *Server) GetLastMovie(ctx context.Context, req *pb.GetLastMovieRequest) (*pb.Movie, error) {
	movie, err := s.uc.GetLastMovie(ctx)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrMovieNotFound):
			return nil, status.Error(codes.NotFound, "no movies found")
		case errors.Is(err, context.DeadlineExceeded):
			return nil, status.Error(codes.DeadlineExceeded, "timeout")
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}

	}

	return &pb.Movie{
		Id:       int32(movie.ID),
		Title:    movie.Title,
		Year:     int32(movie.Year),
		Director: movie.Director,
	}, nil
}
