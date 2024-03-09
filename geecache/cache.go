package geecache

import (
	"geecache/lru"
	"sync"
	"time"
)

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	expiration time.Duration // TTL
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value, time.Now().Add(c.expiration).Unix())
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, t, ok := c.lru.Get(key); ok {
		if time.Now().Unix() > t { // 过期
			c.lru.RemoveKey(key)
		} else {
			return v.(ByteView), ok
		}
	}
	return
}
