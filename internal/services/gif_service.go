package services

import (
	"context"
	"fmt"

	"github.com/pipigendut/dating-backend/internal/providers/gif"
)

type GifService interface {
	Search(ctx context.Context, userID, query, locale string, limit, offset int) ([]gif.Gif, error)
	Trending(ctx context.Context, userID, locale string, limit, offset int) ([]gif.Gif, error)
	GetActiveProvider() string
}

type gifService struct {
	provider gif.Provider
}

// NewGifService injects the active provider determined by configuration
func NewGifService(provider gif.Provider) GifService {
	return &gifService{
		provider: provider,
	}
}

func (s *gifService) Search(ctx context.Context, userID, query, locale string, limit, offset int) ([]gif.Gif, error) {
	if s.provider == nil {
		return nil, fmt.Errorf("gif provider not configured")
	}
	return s.provider.Search(ctx, userID, query, locale, limit, offset)
}

func (s *gifService) Trending(ctx context.Context, userID, locale string, limit, offset int) ([]gif.Gif, error) {
	if s.provider == nil {
		return nil, fmt.Errorf("gif provider not configured")
	}
	return s.provider.Trending(ctx, userID, locale, limit, offset)
}

func (s *gifService) GetActiveProvider() string {
	if s.provider != nil {
		return s.provider.Name()
	}
	return "none"
}
