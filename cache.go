package microgate

import (
	"github.com/cheshir/ttlcache"
	"go.uber.org/zap"
	"time"
)

var cache *LimeCache

type LimeCache struct {
	Cache  *ttlcache.Cache
	Logger *zap.Logger
}

func Cache(totalStorageTime time.Duration) *LimeCache {
	if cache == nil {
		cache := &LimeCache{}
		cache.Cache = ttlcache.New(totalStorageTime)
	}
	return cache
}

func (cache *LimeCache) AttachLogger(logger *zap.Logger) {
	if cache == nil {
		logger.Error("Cache not initiated yet. AttachLogger not possible ...")
	}
	cache.Logger = logger
	logger.Debug("Added Logger to Cache")
}

func (cache *LimeCache) Set(key string, value interface{}, time time.Duration) {
	cache.Cache.Set(ttlcache.StringKey(key), value, time)
	logger.Debug("Added entry to cache", zap.String("key", key), zap.Duration("ttl", time))
}

func (cache *LimeCache) Get(key string) (interface{}, bool) {
	result, ok := cache.Cache.Get(ttlcache.StringKey(key))
	if !ok {
		logger.Warn("Entry not found in cache", zap.String("key", key))
	}
	return result, ok
}
