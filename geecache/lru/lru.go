package lru

import (
	"container/list"
)

type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        *list.List               // 双向链表
	cache     map[string]*list.Element // 哈希表
	OnEvicted func(key string, value Value)
}

type entry struct {
	key        string
	value      Value
	expiration int64 // 过期时间的 Unix() 时间戳
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 查找
func (c *Cache) Get(key string) (value Value, expirationm int64, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, kv.expiration, ok

	}
	return
}

// 删除
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) RemoveKey(key string) {
	if ele, ok := c.cache[key]; ok {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, key)
		c.nbytes -= int64(len(key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(key, kv.value)
		}
	}
}

// 新增/修改
func (c *Cache) Add(key string, value Value, expiration int64) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)

		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		for c.maxBytes != 0 && c.maxBytes < c.nbytes {
			c.RemoveOldest()
		}

		kv.value = value
		kv.expiration = expiration
	} else {
		c.nbytes += int64(len(key)) + int64(value.Len())
		for c.maxBytes != 0 && c.maxBytes < c.nbytes {
			c.RemoveOldest()
		}

		ele := c.ll.PushFront(&entry{key, value, expiration})
		c.cache[key] = ele
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
