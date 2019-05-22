package dcache

import (
	"github.com/hqbobo/log"
	"sync"
	"time"
)

type sobj struct {
	obj string
	ttl int64
}

type MemCache struct {
	cache map[string]sobj
	lock  *sync.RWMutex
}

func newMemCache() *MemCache {
	s := new(MemCache)
	s.cache = make(map[string]sobj, 0)
	s.lock = new(sync.RWMutex)
	return s
}

func (this *MemCache) Get(key string, data *string) bool {
	this.lock.RLock()
	if v, ok := this.cache[key]; ok {
		//log.Debug("--->get mem:", key)
		if time.Now().Unix() < v.ttl {
			this.lock.RUnlock()
			*data = v.obj
			return true
		}
	}
	this.lock.RUnlock()
	return false
}

func (this *MemCache) Set(key string, data string, ttl int) bool {
	this.lock.Lock()
	if _, ok := this.cache[key]; ok {
		delete(this.cache, key)
	}
	o := new(sobj)
	//超时最大为一小时
	if ttl > 60*60 {
		ttl = 60 * 60
	}
	o.ttl = time.Now().Unix() + int64(ttl)
	o.obj = data
	this.cache[key] = *o
	this.lock.Unlock()
	//log.Debug("----->set mem:", key)
	return true
}

func (this *MemCache) Delete(key string) bool {
	this.lock.Lock()
	if _, ok := this.cache[key]; ok {
		log.Debug("----->Delete mem:", key)
		delete(this.cache, key)
	}
	this.lock.Unlock()
	return true
}

func (this *MemCache) Check(key string) bool {
	this.lock.Lock()
	defer this.lock.Unlock()

	if v, ok := this.cache[key]; ok {
		if time.Now().Unix() < v.ttl {
			return true
		} else {
			delete(this.cache, key)
			return false
		}
	}
	return false
}
