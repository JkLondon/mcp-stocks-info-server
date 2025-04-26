package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/patrickmn/go-cache"
)

// InMemoryCache реализация кэша на основе go-cache (in-memory)
type InMemoryCache struct {
	client *cache.Cache
}

// NewInMemoryCache создает новый экземпляр in-memory кэша
func NewInMemoryCache(defaultTTL time.Duration) *InMemoryCache {
	// Создаем кэш с указанным TTL по умолчанию и очисткой устаревших элементов каждые 10 минут
	c := cache.New(defaultTTL, 10*time.Minute)
	return &InMemoryCache{
		client: c,
	}
}

// Get получает значение из кэша
func (c *InMemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, found := c.client.Get(key)
	if !found {
		return nil
	}

	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Set сохраняет значение в кэш
func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.client.Set(key, value, ttl)
	return nil
}

// Delete удаляет значение из кэша
func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
	c.client.Delete(key)
	return nil
}

// Exists проверяет существование ключа в кэше
func (c *InMemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	_, found := c.client.Get(key)
	return found, nil
}

// Invalidate удаляет все ключи соответствующие шаблону
func (c *InMemoryCache) Invalidate(ctx context.Context, pattern string) error {
	// Для простой реализации просто сравниваем начало ключа с шаблоном
	// Более сложная реализация может использовать regexp
	items := c.client.Items()
	for k := range items {
		if len(k) >= len(pattern) && k[:len(pattern)] == pattern {
			c.client.Delete(k)
		}
	}
	return nil
}
