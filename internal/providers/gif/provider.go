package gif

import "context"

// Gif represents a unified GIF structure across different providers
type Gif struct {
	ID       string `json:"id"`
	URL      string `json:"url"` // Media URL (e.g., MP4/GIF actual file)
	Preview  string `json:"preview"` // Thumbnail/preview
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Provider string `json:"provider"`
}

// Provider interface defines the contract for any GIF source (Klipy, Giphy, Tenor, etc.)
type Provider interface {
	Search(ctx context.Context, userID, query, locale string, limit, offset int) ([]Gif, error)
	Trending(ctx context.Context, userID, locale string, limit, offset int) ([]Gif, error)
	Name() string
}
