package cronjob

import (
	"context"
	"sync"
	"time"
)

type Handler func(*Context)

type Context struct {
	context.Context

	mu           sync.RWMutex
	m            map[string]any
	index        int
	handlerChain []Handler
}

func (c *Context) reset(ctx context.Context) *Context {
	c.index = -1
	c.m = nil
	if ctx == nil {
		c.Context = context.Background()
	} else {
		c.Context = ctx
	}

	return c
}

func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[key]
	return v, ok
}

func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.m == nil {
		c.m = make(map[string]any)
	}
	c.m[key] = value
}

func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlerChain) {
		c.handlerChain[c.index](c)
		c.index++
	}
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *Context) Done() <-chan struct{} {
	return nil
}

func (c *Context) Err() error {
	return nil
}

func (c *Context) Value(key any) any {
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}

	return c.Context.Value(key)
}
