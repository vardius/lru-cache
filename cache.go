package lrucache

import (
	"sync"
)

// Cache provides set and get functionality
type Cache interface {
	// Sets value to cache
	Set(id string, value interface{})
	// Gets value from cache
	Get(id string) interface{}
}

type item struct {
	value      interface{}
	previousID string
	nextID     string
}

type items map[string]*item

type cache struct {
	size  int
	items items
	mtx   sync.RWMutex

	lastID  string
	firstID string
}

// New creates new cache
func New(size int) Cache {
	return &cache{
		size:  size,
		items: make(items, size),
	}
}

func (c *cache) Get(id string) interface{} {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	i, exist := c.items[id]
	if !exist {
		return nil
	}

	c.moveItemToTheHead(id, i)

	return i.value
}

func (c *cache) Set(id string, value interface{}) {
	if id == "" {
		return
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	i, exist := c.items[id]
	if !exist {
		if c.isFull() {
			c.removeLastItem()
		}

		i = &item{
			value:      value,
			previousID: c.firstID,
		}
	}

	c.moveItemToTheHead(id, i)
}

func (c *cache) removeLastItem() {
	lastID := c.lastID
	c.lastID = c.items[lastID].nextID

	if c.lastID != "" {
		c.items[c.lastID].previousID = ""
	}

	delete(c.items, lastID)
}

func (c *cache) moveItemToTheHead(id string, i *item) {
	if c.lastID == "" {
		c.lastID = id
	}

	if i.previousID != "" && i.previousID != c.lastID {
		c.items[i.previousID].nextID = i.nextID
	}

	if i.nextID != "" {
		c.items[i.nextID].previousID = i.previousID
	}

	if c.firstID != "" {
		c.items[c.firstID].nextID = id
	}

	i.previousID = c.firstID
	i.nextID = ""

	c.firstID = id
	c.items[id] = i
}

func (c *cache) isFull() bool {
	return len(c.items) == c.size
}
