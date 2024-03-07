package singleflight

import "sync"

// call 表示正在进行中或已完成的 Do 调用
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 表示一类工作，并形成一个命名空间，在此空间内，
// 可以执行带有重复抑制机制的工作单元。
type Group struct {
	mu sync.Mutex       // 锁
	m  map[string]*call // 初始化
}

// Do 执行给定的函数并返回其结果，确保在给定键下同时只有一个执行实例在进行中。
// 如果有重复调用进入，重复调用者会等待原始调用完成，并接收相同的执行结果。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
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
