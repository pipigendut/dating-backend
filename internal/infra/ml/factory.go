package ml

import (
	"fmt"
	"os"
)

// FaceVerificationFactory returns the configured provider based on an environment variable
func NewProvider() (FaceVerificationProvider, error) {
	providerType := os.Getenv("FACE_VERIFICATION_PROVIDER")
	if providerType == "" {
		providerType = "mock" // Default to mock for development
	}

	switch providerType {
	case "mock":
		return NewMockProvider(), nil
	case "aws":
		// return NewAWSProvider(), nil
		return nil, fmt.Errorf("aws provider not yet implemented")
	case "onnx":
		modelPath := os.Getenv("ONNX_MODEL_PATH")
		if modelPath == "" {
			modelPath = "./models/arcface_resnet50.onnx"
		}
		libPath := os.Getenv("ONNXRUNTIME_SHARED_LIBRARY_PATH")
		return NewONNXProvider(modelPath, libPath)
	default:
		return NewMockProvider(), nil
	}
}
