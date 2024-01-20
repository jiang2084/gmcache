package lru

import (
	"container/list"
)

/*
1. 字典是存储键值的映射关系，根据某个key查询值的时间复杂度是o(1),在字典中插入一条记录的时间复杂度也是o(1)
2. 所有的值放在双向链表里面，访问到某个值时将其移动到队尾的时间复杂度是o(1)
*/

type Cache struct {
	maxBytes  int64      // 最大允许使用内存大小
	nBytes    int64      // 当前使用内存大小
	ll        *list.List // 使用go自带的list双向列表
	cache     map[string]*list.Element
	OnEvicted func(keys string, value Value)
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//  功能实现

// Get 1. 查找
func (c *Cache) Get(key string) (value Value, ok bool) {
	if element, ok := c.cache[key]; ok {
		// 如果对应的元素存在，则将对应节点移动到队头，并返回查找的值
		// 使用过的元素移动到队头
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 2. 超过容量，删除部分缓存
func (c *Cache) RemoveOldest() {
	// 约定对头的元素为最近使用的，越是队尾的元素越是最早访问的，删除即可
	element := c.ll.Back()
	if element != nil {
		c.ll.Remove(element)
		kv := element.Value.(*entry)
		delete(c.cache, kv.key)
		// 删除的key 需要减去键的长度和值的长度
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 3. 插入
func (c *Cache) Add(key string, value Value) {
	// 这个key之前存在，需要更新值，并移动位置
	if element, ok := c.cache[key]; ok {
		c.ll.MoveToFront(element)
		kv := element.Value.(*entry)
		// 减去之前value的长度，加上本次的长度
		c.nBytes = c.nBytes - int64(kv.value.Len()) + int64(value.Len())
	} else {
		// 不存在，则直接插入到队首即可
		element := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = element
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	// 超过长度了，删除之前的
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// Len 获取总的缓存个数
func (c *Cache) Len() int {
	return c.ll.Len()
}
