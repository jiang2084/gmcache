package gmcache

import (
	"hash/crc32"
	"sort"
	"strconv"
)

/*
一致性哈希
1. 解决的问题：解决简单哈希中，节点数量变化的场景；简单哈希中， 如果某个节点挂掉了，节点数量变化后
几乎导致所有缓存值对应的节点都会发生改变，即几乎所有缓存值都失效了。
节点接收到数据均需要重新去数据源获取数据，容易发生雪崩

缓存雪崩：缓存在同一时刻全部失效，造成瞬时DB请求量大，压力骤增，引起雪崩。

2. 原理
* 计算节点的哈希值放在环上
* 计算key的哈希值，放置在环上，顺时针寻找到第一个节点，就是应选取的节点

优点：一致性哈希算法，在新增、删除节点时，只需要重新定义该节点附近的一小部分数据
而不需要重新定位所有的节点

缺点：出现数据倾斜问题；解决：引入虚拟节点
*/

type Hash func(data []byte) uint32

type Map struct {
	hash    Hash
	replica int // 虚拟节点倍数
	keys    []int
	hashMap map[int]string // 虚拟节点和真实节点的映射表
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replica: replicas,
		hash:    fn,
		hashMap: make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 实现添加真实节点/机器的Add()方法
/*
   通过虚拟哈希，可能存在多个哈希节点对应一个真实节点的情况
*/
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replica; i++ {
			// []byte(strconv.Itoa(i) + key) 这里就是做了字符串的拼接
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	// 哈希环排序
	sort.Ints(m.keys)
}

// Get 实现节点的Get()方法
func (m *Map) Get(key string) string {
	//获取哈希中与提供的键最接近的项。
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个匹配的虚拟节点的下标，环状结构所以应该取模
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
