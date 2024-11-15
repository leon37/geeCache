package logic

import (
	"errors"
	pb "geeCache/logic/geecachepb"
	"geeCache/logic/singleflight"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func Register(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("geeCache: Register getter is nil")
	}
	mu.Lock()
	defer mu.Unlock()
	group := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    new(singleflight.Group),
	}
	groups[name] = group
	return group
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}

	value, ok := g.mainCache.get(key)
	if ok {
		log.Println("[geeCache] hit")
		return value, nil
	} else {
		return g.load(key)
	}
}

func (g *Group) load(key string) (ByteView, error) {
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return view.(ByteView), nil
	}
	return ByteView{}, err
}

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

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers == nil {
		panic("Register peerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	rsp := &pb.Response{}
	err := peer.Get(req, rsp)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: rsp.GetValue()}, nil
}
