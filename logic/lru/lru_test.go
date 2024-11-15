package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(0, nil)
	lru.Add("key1", String("value1"))
	if v, ok := lru.Get("key1"); !ok || v.(String) != "value1" {
		t.Fatalf("cache hit key1=value1 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestAdd(t *testing.T) {
	lru := New(0, nil)
	lru.Add("key1", String("value1"))
	lru.Add("key1", String("value22"))
	if lru.currentBytes != int64(len("key1")+len("value22")) {
		t.Fatalf("cache update key1 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	capacity := len(k1 + k2 + v1 + v2)
	lru := New(int64(capacity), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k1, String(v1))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key2"); ok || lru.Len() != 2 {
		t.Fatalf("cache update key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
