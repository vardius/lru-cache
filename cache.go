package lrucache

import (
	"container/list"
	"fmt"
	"sync"
)

type ByteSize uint64

const (
	_           = iota
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

// Cache provides set and get functionality
type Cache interface {
	// Sets value to cache
	Set(key string, value []byte) error
	// Gets value from cache
	Get(key string) ([]byte, error)
}

type item struct {
	key   string
	value []byte
}

type cache struct {
	list     *list.List
	elements map[string]*list.Element
	mtx      sync.Mutex

	name        string
	maxSize     ByteSize
	currentSize ByteSize
}

// New creates new cache
func New(name string, maxSize ByteSize) (Cache, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	c := &cache{
		list: list.New(),

		name:     name,
		maxSize:  maxSize,
		elements: make(map[string]*list.Element),
	}

	return c, nil
}

func (c *cache) Get(key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("empty key")
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if element, exist := c.elements[key]; exist {
		c.list.MoveToFront(element)

		return element.Value.(item).value, nil
	}

	return nil, nil
}

func (c *cache) Set(key string, value []byte) error {
	if key == "" {
		return fmt.Errorf("empty key")
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	if element, ok := c.elements[key]; ok {
		c.list.MoveToFront(element)

		item := element.Value.(item)
		oldItemSize := len(item.value)
		newItemSize := len(value)

		c.currentSize -= ByteSize(oldItemSize)
		c.currentSize += ByteSize(newItemSize)

		item.value = value
	} else {
		itemSize := len(value)

		c.currentSize += ByteSize(itemSize)
		c.elements[key] = c.list.PushFront(item{
			key:   key,
			value: value,
		})

		for c.currentSize >= c.maxSize {
			c.removeLastItem()
		}
	}

	return nil
}

func (c *cache) removeLastItem() {
	element := c.list.Back()
	item := element.Value.(item)
	itemSize := len(item.value)

	delete(c.elements, item.key)

	c.list.Remove(element)
	c.currentSize -= ByteSize(itemSize)
}
