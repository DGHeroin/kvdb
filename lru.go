package kvdb

import (
    "container/list"
    "sync"
)

const (
    RemoveTypeFullEntries = RemoveReason(0)
    RemoveTypeFullMemory  = RemoveReason(1)
    RemoveTypeByUser      = RemoveReason(2)
)
type Cache interface {
    Add(key []byte, value []byte)
    Get(key []byte) (value []byte, ok bool)
    Remove(key []byte)
    RemoveOldest()
    Len() uint64
    Clear()
}

type lru struct {
    MaxEntries  uint64
    MaxMemory   uint64
    memoryCount uint64
    OnEvicted   func(key []byte, value []byte, fullType RemoveReason)
    ll          *list.List
    cache       map[string]*list.Element
    mu          sync.RWMutex
}

type RemoveReason int

type entry struct {
    key   []byte
    value []byte
}

func NewCache(maxEntries uint64, maxMemory uint64) Cache {
    if maxEntries == 0 {
        maxEntries = ^uint64(0)
    }
    if maxMemory == 0 {
        maxMemory = ^uint64(0)
    }

    return &lru{
        MaxEntries: maxEntries,
        MaxMemory:  maxMemory,
        ll:         list.New(),
        cache:      make(map[string]*list.Element),
    }
}

func (c *lru) Add(key []byte, value []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()
    keyStr := string(key)
    if c.cache == nil {
        c.cache = make(map[string]*list.Element)
        c.ll = list.New()
    }
    itemSize := uint64(len(value))
    // check memory enough
    if itemSize > c.MaxMemory {
        // too large
        return
    }

    for {
        left := c.MaxMemory - c.memoryCount
        if c.lenLocked() == 0 {
            // remove all but memory still not enough
            if left > itemSize {
                break
            }
            return
        }
        if left < itemSize {
            c.removeOldestLocked(RemoveTypeFullMemory)
            continue
        }
        break
    }

    if ee, ok := c.cache[keyStr]; ok {
        c.ll.MoveToFront(ee)
        ee.Value.(*entry).value = value
        c.memoryCount += itemSize
        return
    }
    c.memoryCount += itemSize
    ele := c.ll.PushFront(&entry{key, value})
    c.cache[keyStr] = ele
    if uint64(c.ll.Len()) > c.MaxEntries {
        c.removeOldestLocked(RemoveTypeFullEntries)
    }
}

func (c *lru) Get(key []byte) (value []byte, ok bool) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.cache == nil {
        return
    }
    keyStr := string(key)
    if ele, hit := c.cache[keyStr]; hit {
        c.ll.MoveToFront(ele)
        return ele.Value.(*entry).value, true
    }
    return
}

func (c *lru) Remove(key []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()
    keyStr := string(key)
    if c.cache == nil {
        return
    }
    if ele, hit := c.cache[keyStr]; hit {
        c.removeElement(ele, RemoveTypeByUser)
    }
}

func (c *lru) RemoveOldest() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.removeOldestLocked(RemoveTypeByUser)
}

func (c *lru) removeOldestLocked(removeType RemoveReason) {
    if c.cache == nil {
        return
    }
    ele := c.ll.Back()
    if ele != nil {
        if data, ok := ele.Value.(*entry); ok {
            c.memoryCount -= uint64(len(data.value))
        }
        c.removeElement(ele, removeType)
    }
}

func (c *lru) removeElement(e *list.Element, removeType RemoveReason) {
    c.ll.Remove(e)
    kv := e.Value.(*entry)
    keyStr := string(kv.key)
    delete(c.cache, keyStr)
    if c.OnEvicted != nil {
        c.OnEvicted(kv.key, kv.value, removeType)
    }
}

func (c *lru) Len() uint64 {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.lenLocked()
}

func (c *lru) lenLocked() uint64 {
    if c.cache == nil {
        return 0
    }
    return uint64(c.ll.Len())
}

func (c *lru) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.OnEvicted != nil {
        for _, e := range c.cache {
            kv := e.Value.(*entry)
            c.OnEvicted(kv.key, kv.value, RemoveTypeFullEntries)
        }
    }
    c.ll = nil
    c.cache = nil
}

func (r RemoveReason) String() string {
    switch r {
    case RemoveTypeFullEntries:
        return "Remove by full entries"
    case RemoveTypeFullMemory:
        return "Remove by full memory"
    case RemoveTypeByUser:
        return "Remove by user"
    }
    return "Unknown remove reason"
}

