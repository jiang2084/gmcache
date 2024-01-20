package lru

import (
	"reflect"
	"testing"
)

type String string

// 显示接口声明

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	// 容量为0，说明缓存就没有添加进去
	lru.Add("k1", String("1234"))

	if _, ok := lru.Get("k1"); !ok {
		t.Fatalf("cache hit fail")
	}

	if v, ok := lru.Get("k1"); !ok || string(v.(String)) == "1234" {
		t.Fatalf("value is wrong")
	}

	if _, ok := lru.Get("k2"); !ok {
		t.Fatalf("cache miss k2 fail")
	}
}

func TestRemoveoldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	size := len(k1 + k2 + v1 + v2)
	lru := New(int64(size), nil)

	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get(k1); ok || lru.Len() == 2 {
		t.Fatalf("not found")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lru := New(int64(10), callback)
	lru.Add("key1", String("123456")) // 4 + 6
	lru.Add("k2", String("k2"))       // 2 + 2  删除key1
	lru.Add("k3", String("k3"))       // 2 + 2
	lru.Add("k4", String("k4"))       // 2 + 2  删除 k2

	expect := []string{"key1", "k2"}
	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
