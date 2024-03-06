package lru

import "container/list"

type Cache struct {
	maxBytes  int64
	nbytes    int64
	ll        *list.List               // 双向链表
	cache     map[string]*list.Element // 哈希表
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
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
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true

	}
	return
}

//删除
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

// 新增/修改
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)

		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		for c.maxBytes != 0 && c.maxBytes < c.nbytes {
			c.RemoveOldest()
		}

		kv.value = value
	} else {
		c.nbytes += int64(len(key)) + int64(value.Len())
		for c.maxBytes != 0 && c.maxBytes < c.nbytes {
			c.RemoveOldest()
		}

		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
