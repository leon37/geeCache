package logic

import (
	"errors"
	"log"
	"reflect"
	"testing"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("expect %v, got %v callback failed", expect, v)
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	gee := Register("Scores", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			loadCounts[key]++
			return []byte(v), nil
		}
		return nil, errors.New("key not found")
	}))

	for k, v := range db {
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %v", k)
		}
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %v miss", k)
		}
	}

	if value, err := gee.Get("unknown"); err == nil {
		t.Fatalf("get value of unknown not found %v", value)
	}
}

func TestGetGroup(t *testing.T) {
	groupName := "scores"
	Register(groupName, 2<<10, GetterFunc(
		func(key string) (bytes []byte, err error) { return }))
	if group := GetGroup(groupName); group == nil || group.name != groupName {
		t.Fatalf("group %s not exist", groupName)
	}

	if group := GetGroup(groupName + "111"); group != nil {
		t.Fatalf("expect nil, but %s got", group.name)
	}
}
