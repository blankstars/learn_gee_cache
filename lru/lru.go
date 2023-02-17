package lru

// lru缓存淘汰策略

import "container/list"

type Cache struct {
	maxBytes  int64                         // 最大内存
	nbytes    int64                         // 当前内存
	ll        *list.List                    // 存储内容的双向列表
	cache     map[string]*list.Element      // k-v对，key为键值, value为双向列表中的元素
	OnEvicted func(key string, value Value) // 元素被驱逐时回调
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nbytes:    0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // 更新
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(kv.value.Len()) - int64(value.Len())
		kv.value = value
	} else { // 新增
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
