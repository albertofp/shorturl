package server

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/albertofp/shorturl/api/proto/shorturl/v1"
)

// URLShortenerServer is the server implementation for the URLShortener gRPC service.
type URLShortenerServer struct {
	pb.UnimplementedURLShortenerServer
	DB URLDatabase
}

type URLDatabase interface {
	// SaveShortURL stores a new short URL in the database.
	SaveShortURL(ctx context.Context, shortURL, longURL string, createdAt time.Time, ttl time.Time) (int64, error)

	// GetLongURL retrieves the original long URL by its short URL ID.
	GetLongURL(ctx context.Context, shortURL string) (longURL string, createdAt time.Time, ttl time.Time, err error)
}

// CreateShortURL handles the creation of a shortened URL.
func (s *URLShortenerServer) CreateShortURL(ctx context.Context, req *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	// TODO: base url
	createdAt := time.Now()
	shortURL := s.genShortURL("https://shorturl.albertofp.com", req.Longurl)

	// TODO: customize ttl
	id, err := s.DB.SaveShortURL(ctx, shortURL, req.Longurl, createdAt, createdAt.Add(48*time.Hour))
	if err != nil {
		slog.Error("Failed to save short URL", slog.String("error", err.Error()))
		return nil, err
	}

	slog.Info("Created short URL",
		slog.Int64("id", id),
		slog.String("shortURL", shortURL),
		slog.String("longURL", req.Longurl),
		slog.String("createdAt", createdAt.Format(time.RFC3339)),
	)
	// TODO: cache

	return &pb.CreateShortURLResponse{
		Id:        id,
		Shorturl:  shortURL,
		CreatedAt: createdAt.Format(time.RFC3339),
	}, nil
}

func (s *URLShortenerServer) GetLongURL(ctx context.Context, req *pb.GetLongURLRequest) (*pb.GetLongURLResponse, error) {
	// Use the GetLongURL method of the database interface
	longURL, createdAt, ttl, err := s.DB.GetLongURL(ctx, req.Shorturl)
	if err != nil {
		slog.Error("Failed to retrieve long URL", slog.String("error", err.Error()), slog.String("shorturl", req.Shorturl))
		return nil, err
	}

	// Convert Go time.Time to Protobuf Timestamp
	return &pb.GetLongURLResponse{
		Longurl:   longURL,
		CreatedAt: timestamppb.New(createdAt),
		Ttl:       timestamppb.New(ttl),
	}, nil
}

func (s *URLShortenerServer) genShortURL(baseURL, longURL string) string {
	hash := md5.Sum([]byte(longURL))
	encodedHash := base64.URLEncoding.EncodeToString(hash[:])

	return strings.Join([]string{baseURL, encodedHash}, "/")
}
