package lru

import "container/list"

type Cache struct {
	maxBytes     int64
	currentBytes int64
	list         *list.List
	mp           map[string]*list.Element
	OnEvicted    func(key string, value Value)
}

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

func New(maxBytes int64, f func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		list:      list.New(),
		mp:        make(map[string]*list.Element),
		OnEvicted: f,
	}
}

func (c *Cache) Len() int {
	return c.list.Len()
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if e, ok := c.mp[key]; ok {
		c.list.MoveToBack(e)
		value = e.Value.(*entry).value
		return value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	e := c.list.Front()
	if e != nil {
		c.list.Remove(e)
		pair := e.Value.(*entry)
		delete(c.mp, pair.key)
		c.currentBytes -= int64(len(pair.key)) + int64(pair.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(pair.key, pair.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	e, ok := c.mp[key]
	if !ok {
		e = c.list.PushBack(&entry{key, value})
		c.mp[key] = e
		c.currentBytes += int64(len(key)) + int64(value.Len())
	} else {
		c.list.MoveToBack(e)
		pair := e.Value.(*entry)
		c.currentBytes += int64(value.Len()) - int64(pair.value.Len())
		pair.value = value
	}
	for c.maxBytes > 0 && c.currentBytes > c.maxBytes {
		c.RemoveOldest()
	}
}
