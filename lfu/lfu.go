package lfu

import (
	"container/list"
)

type Cache struct {
	maxBytes   int64
	nbytes     int64
	minFreq    int
	cache      map[string]*list.Element // 哈希表
	freqToList map[int]*list.List       // 使用频率
	OnEvicted  func(key string, value Value)
}

type entry struct {
	key   string
	value Value
	freq  int
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:   maxBytes,
		freqToList: make(map[int]*list.List),
		cache:      make(map[string]*list.Element),
		OnEvicted:  onEvicted,
	}
}

// 查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		lst := c.freqToList[kv.freq]
		lst.Remove(ele)
		if lst.Len() == 0 {
			delete(c.freqToList, kv.freq)
			if c.minFreq == kv.freq {
				c.minFreq++
			}
		}

		kv.freq++
		if _, ok := c.freqToList[kv.freq]; !ok {
			c.freqToList[kv.freq] = list.New()
		}
		c.cache[kv.key] = c.freqToList[kv.freq].PushFront(kv)
		return kv.value, true
	}
	return
}

// 删除
func (c *Cache) RemoveOldest() {
	lst := c.freqToList[c.minFreq]
	ele := lst.Back()
	if ele != nil {
		lst.Remove((ele))
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}
