package services

import (
	"context"
	"github.com/pipigendut/dating-backend/internal/repository"
	"log"
	"strconv"
	"sync"
	"time"
)

type ConfigService interface {
	LoadConfigs(ctx context.Context) error
	GetString(key string, defaultVal string) string
	GetInt(key string, defaultVal int) int
	GetFloat(key string, defaultVal float64) float64
	Set(ctx context.Context, key, value string) error
	ResetDB(ctx context.Context) error
	GetAllCached(ctx context.Context) map[string]string
}

type configService struct {
	repo    repository.ConfigRepository
	configs map[string]string
	mu      sync.RWMutex
}

func NewConfigService(repo repository.ConfigRepository) ConfigService {
	srv := &configService{
		repo:    repo,
		configs: make(map[string]string),
	}
	
	// Initial load
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.LoadConfigs(ctx); err != nil {
		log.Printf("Warning: Failed to load initial configs from DB: %v", err)
	}

	return srv
}

func (s *configService) LoadConfigs(ctx context.Context) error {
	configs, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, cfg := range configs {
		s.configs[cfg.Key] = cfg.Value
	}

	return nil
}

func (s *configService) GetString(key string, defaultVal string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if val, ok := s.configs[key]; ok {
		return val
	}
	return defaultVal
}

func (s *configService) GetInt(key string, defaultVal int) int {
	strVal := s.GetString(key, "")
	if strVal == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(strVal)
	if err != nil {
		log.Printf("Warning: Invalid int config value for key %s: %s", key, strVal)
		return defaultVal
	}
	return val
}

func (s *configService) GetFloat(key string, defaultVal float64) float64 {
	strVal := s.GetString(key, "")
	if strVal == "" {
		return defaultVal
	}

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		log.Printf("Warning: Invalid float config value for key %s: %s", key, strVal)
		return defaultVal
	}
	return val
}

func (s *configService) Set(ctx context.Context, key, value string) error {
	if err := s.repo.Set(ctx, key, value); err != nil {
		return err
	}

	s.mu.Lock()
	s.configs[key] = value
	s.mu.Unlock()

	return nil
}

func (s *configService) ResetDB(ctx context.Context) error {
	return s.repo.DeleteAll(ctx)
}

func (s *configService) GetAllCached(ctx context.Context) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range s.configs {
		result[k] = v
	}
	return result
}
