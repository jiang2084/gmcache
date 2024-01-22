package singleflight

import "sync"

/**
1. 缓存雪崩：缓存在同一时间全部失效，造成瞬时DB请求量大，压力骤增，引起雪崩（key全部失效）
2. 缓存击穿：一个存在的key，在缓存过期的一刻同时有大量的请求，这些请求会击穿到DB（某个key失效，大量请求）
3. 缓存穿透：查询一个不存在的key，因为不存在所以不会写缓存，造成每次请求都到DB，（查询一个不在的key）
瞬时流量过大，穿透到DB，导致宕机
*/
// call 代表正在进行中，或者已经结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 方法接收2个参数，key和func，针对相同的key，无论Do被调用多少次，函数fn都只会被调用一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // 如果请求正在进行中，则等待
		return c.val, c.err // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
