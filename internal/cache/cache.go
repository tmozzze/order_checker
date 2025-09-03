package cache

// TODO: Доделать кеш c емкостью

import (
	"container/list"
	"sync"

	"github.com/tmozzze/order_checker/internal/models"
)

type entry struct {
	key   string
	value interface{}
}
type Cache struct {
	capacity int
	mu       sync.RWMutex
	ll       *list.List
	store    map[string]*list.Element
}

func New(capacity int) *Cache {
	return &Cache{
		capacity: capacity,
		ll:       list.New(),
		store:    make(map[string]*list.Element),
	}

}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.store[key]; ok {
		c.ll.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	elem := c.ll.PushFront(&entry{key: key, value: value})
	c.store[key] = elem

	if c.ll.Len() > c.capacity {
		back := c.ll.Back()
		if back != nil {
			c.ll.Remove(back)
			delete(c.store, back.Value.(*entry).key)
		}
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if elem, ok := c.store[key]; ok {
		c.ll.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return nil, false
}

func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.store[key]; ok {
		c.ll.Remove(elem)
		delete(c.store, key)
	}

	return nil
}

func (c *Cache) LoadAll(orders []models.Order) {
	for i, order := range orders {
		c.Set(order.OrderUID, order)
		if i == c.capacity {
			return
		}
	}
}

func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}
