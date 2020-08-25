package cache

type Cache struct {
	ch chan interface{}
}

func NewCache(cap int) *Cache {
	c := &Cache{
		ch: make(chan interface{}, cap),
	}
	return c
}

func (c *Cache) Enter(a interface{}) {
	c.ch <- a
}

func (c *Cache) Out() interface{} {
	return <- c.ch
}


