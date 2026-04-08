package mapper

// StorageURLProvider is a shared interface for resolving public URLs
type StorageURLProvider interface {
	GetPublicURL(key string) string
}
