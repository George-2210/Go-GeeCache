package lfu

import (
	"container/list"
	"fmt"
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

func (c *Cache) PushFront(e *entry) {
	if _, ok := c.freqToList[e.freq]; !ok {
		c.freqToList[e.freq] = list.New() // 双向链表
	}
	c.cache[e.key] = c.freqToList[e.freq].PushFront(e)
}

// 删除
func (c *Cache) RemoveOldest() {
	if lst, ok := c.freqToList[c.minFreq]; ok {
		ele := lst.Back()
		if ele != nil {
			lst.Remove(ele)
			kv := ele.Value.(*entry)
			delete(c.cache, kv.key)
			c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
			if c.OnEvicted != nil {
				c.OnEvicted(kv.key, kv.value)
			}
		}
		if lst.Len() == 0 {
			delete(c.freqToList, c.minFreq)
		}
	} else {
		c.minFreq++
		c.RemoveOldest()
	}
}

func (c *Cache) GetEntry(key string) (e *entry, ok bool) {
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
		c.PushFront(kv)
		return kv, true
	}
	return
}

// 查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	if kv, ok := c.GetEntry(key); ok {
		return kv.value, true
	}
	return
}

// 新增/修改
func (c *Cache) Add(key string, value Value) {
	if kv, ok := c.GetEntry(key); ok {

		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		fmt.Println(value, int64(value.Len()))
		fmt.Println(kv.value, int64(kv.value.Len()))
		for c.maxBytes != 0 && c.maxBytes < c.nbytes {
			c.RemoveOldest()
		}

		kv.value = value
		return
	} else {
		// 新增
		c.nbytes += int64(len(key)) + int64(value.Len())
		for c.maxBytes != 0 && c.maxBytes < c.nbytes {
			c.RemoveOldest()
		}

		if _, ok := c.freqToList[1]; !ok {
			c.freqToList[1] = list.New()
		}
		lst := c.freqToList[1]
		ele := lst.PushFront(&entry{key, value, 1})
		c.cache[key] = ele
		c.minFreq = 1

	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}

}

func (c *Cache) Len() int {
	return len(c.cache)
}
