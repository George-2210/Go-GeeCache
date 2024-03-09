package geecache

import (
	"fmt"
	"time"

	// pb "geecache/geecachepb"
	"geecache/singleflight"
	"log"
	"sync"
)

// Group 是一个缓存的命名空间，包含了与数据加载相关的数据
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group
}

// Getter 负责为指定的键加载数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 使用函数实现了 Getter 接口
type GetterFunc func(key string) ([]byte, error)

// Get 实现了 Getter 接口中的 Get 方法
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建一个新的 Group 实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes, expiration: 1 * time.Minute}, // 默认设置5分组，为防止缓存雪崩，可以增加随机数
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup 返回先前使用 NewGroup 创建的命名组，如果没有这样的组，则返回 nil
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 从缓存中获取指定键的值
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// RegisterPeers 用于注册一个 PeerPicker，用于选择远程节点。
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {

	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// 分布式场景 可以调用 getFromPeer 从其他节点获取
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	// req := &pb.Request{
	// 	Group: g.name,
	// 	Key:   key,
	// }
	// res := &pb.Response{}
	// err := peer.Get(req, res)
	// if err != nil {
	// 	return ByteView{}, err
	// }
	// return ByteView{b: res.Value}, nil
	return ByteView{}, nil
}
