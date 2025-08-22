package cache

import (
	"L0/internal/models"
	"errors"
	"sync"
)

type Cache struct {
	mu    sync.RWMutex
	store map[string]models.Order
}

func New() *Cache {
	return &Cache{store: make(map[string]models.Order)}
}

func (c *Cache) Set(id string, o models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[id] = o
}

func (c *Cache) Get(id string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok := c.store[id]

	return v, ok
}

func (c *Cache) Delete(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, found := c.store[id]; !found {
		err := errors.New("Id not found")
		return err
	}

	delete(c.store, id)

	return nil
}
