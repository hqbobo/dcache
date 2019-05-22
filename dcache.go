package dcache

import "lyc-app/src/lib/json"

type Cache interface {
	Check(key string) bool
	CheckMem(key string) bool
	Get(key string, data interface{}) bool
	Set(key string, data interface{}, ttl int) bool
	SetTextSerialize(text TextSerialize)
	Delete(key string) bool
}

type TextSerialize interface {
	Marshal(o interface{}) (string, error)
	Unmarshal(data []byte, v interface{}) error
}

var cache Cache

type defaultText struct {
}

func (this defaultText) Marshal(o interface{}) (string, error) {
	return json.Marshal(o)
}

func (this defaultText) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type Options struct {
	Ip          string
	Port        int
	Pass        string
	Db          int
	PoolSize    int
	ClusterMode bool
}

func Init(option Options) {
	if option.ClusterMode {
		cache = newRedisClusterCache(option.Ip, option.Port, option.Pass, option.Db, option.PoolSize, defaultText{})
	} else {
		cache = newRedisCache(option.Ip, option.Port, option.Pass, option.Db, option.PoolSize, defaultText{})
	}
}

func GetCache() Cache {
	return cache
}
