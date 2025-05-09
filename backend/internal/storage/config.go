package storage

// StorageConfig holds configuration for file storage
type StorageConfig struct {
    Type      string // "local" or "s3"
    LocalPath string // for local storage
    BaseURL   string // base URL for serving files
}

// NewLocalStorageConfig returns default local storage configuration
func NewLocalStorageConfig() *StorageConfig {
    return &StorageConfig{
        Type:      "local",
        LocalPath: "uploads",
        BaseURL:   "/uploads",
    }
} 