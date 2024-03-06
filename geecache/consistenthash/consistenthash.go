package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 将字节映射到 uint32
type Hash func(data []byte) uint32

// Map 包含所有已哈希映射的键
type Map struct {
	hash     Hash  //哈希函数
	replicas int   // 虚拟节点倍数
	keys     []int // 哈希环
	hashMap  map[int]string
}

// New 创建一个 Map 实例
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 向哈希中添加一些真实节点。
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 获取哈希中最接近提供的键的项。
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 二分查找适当的副本。
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
