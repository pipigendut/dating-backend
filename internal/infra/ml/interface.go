package ml

import "context"

// FaceVerificationProvider defines the interface for different face matching engines (ONNX, AWS, etc.)
type FaceVerificationProvider interface {
	// CompareFaces compares two face images and returns a similarity score (0.0 to 1.0)
	CompareFaces(ctx context.Context, face1, face2 []byte) (float64, error)
	// Close cleans up resources (like the ONNX runtime environment)
	Close() error
}

// VerificationResult represents the outcome of a comparison
type VerificationResult struct {
	IsMatch    bool    `json:"is_match"`
	Confidence float64 `json:"confidence"`
	Error      string  `json:"error,omitempty"`
}
