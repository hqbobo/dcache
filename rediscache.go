package dcache

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/hqbobo/log"
	"libs/encrypt"
	"time"
)

const (
	redis_item_timeout = 60 * 60
	redis_sync_chan    = "sync"
	redis_sync_set     = 1
	redis_sync_del     = 2
)

type publisher struct {
	From string
	Act  int
	Key  string
	Val  string
	Ttl  int
}

type RedisCache struct {
	pool   *redis.Pool
	ip     string
	port   int
	pass   string
	db     int
	name   string
	mem    *MemCache
	text   TextSerialize
	logger Logger
}

func newRedisCache(ip string, port int, pass string, db int, poolsize int, text TextSerialize, logger Logger) *RedisCache {
	s := new(RedisCache)
	s.ip = ip
	s.pass = pass
	s.port = port
	s.db = db
	s.name = encrypt.GetRandomString(16)
	s.logger = logger
	s.pool = &redis.Pool{
		// Other pool configuration not shown in this example.
		MaxIdle:     poolsize,
		MaxActive:   poolsize,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
			if err != nil {
				log.Error(err)
				return nil, err
			}
			if len(pass) > 0 {
				if _, err := c.Do("AUTH", pass); err != nil {
					log.Error(err)
					c.Close()
					return nil, err
				}
			}
			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
	s.text = text
	s.mem = newMemCache()
	go s.subscribe()
	return s
}

//监听数据修改事件
func (this *RedisCache) subscribe() {
	c := this.pool.Get()
	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe(redis_sync_chan)
	var pub publisher
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			if e := this.text.Unmarshal(v.Data, &pub); e == nil {
				log.Debug(v.Channel, "[", pub.From, "]:message:", string(v.Data))
				if pub.From != this.name {
					log.Debug("it from others")
					if pub.Act == redis_sync_set {
						this.mem.Set(pub.Key, pub.Val, pub.Ttl)
					} else if pub.Act == redis_sync_del {
						this.mem.Delete(pub.Key)
					}
				}
			} else {
				log.Warn(e)
			}
		case redis.Subscription:
			log.Debug(v.Channel, ":", v.Kind, ":", v.Count)
		case error:
			return
		}
	}
}

//消息推送
func (this *RedisCache) publish(key, val string, ttl int, act int) {
	p := new(publisher)
	p.Key = key
	p.Val = val
	p.Ttl = ttl
	p.Act = act
	p.From = this.name
	// 从池里获取连接
	conn := this.pool.Get()
	defer conn.Close()
	if conn == nil {
		log.Warn("failed to get reids conn")
		return
	}

	//转为字符串
	s, e := this.text.Marshal(p)
	if e != nil {
		log.Warn(e)
		return
	}

	r, e := redis.Int(conn.Do("PUBLISH", redis_sync_chan, s))
	if e != nil || r < 0 {
		log.Warn(e)
	}
}

//获取超时
func (this *RedisCache) getTtl(key string) (int, bool) {
	conn := this.pool.Get()
	i, e := redis.Int(conn.Do("ttl", key))
	defer conn.Close()
	if e != nil {
		log.Warn(e)
		return -1, false
	}
	return i, true
}

func (this *RedisCache) Get(key string, data interface{}) bool {
	var s string
	var e error
	if !this.mem.Get(key, &s) {
		// 从池里获取连接
		conn := this.pool.Get()
		s, e = redis.String(conn.Do("Get", key))
		defer conn.Close()
		if e != nil {
			log.Warn(e)
			return false
		}
		if ttl, ok := this.getTtl(key); ok {
			log.Debug("load:", key, " ttl[", ttl, "] from redis:")
			//设置本地内存
			if e := this.text.Unmarshal([]byte(s), data); e != nil {
				log.Warn(e)
				return false
			}
			//内存提前5秒超时
			return this.mem.Set(key, s, ttl-5)
		}
		return false
	}
	if e := this.text.Unmarshal([]byte(s), data); e != nil {
		log.Warn(e)
		return false
	}
	return true
}

func (this *RedisCache) Set(key string, data interface{}, ttl int) bool {
	// 从池里获取连接
	conn := this.pool.Get()
	if conn == nil {
		return false
	}
	defer conn.Close()

	//转为字符串
	s, e := this.text.Marshal(data)
	if e != nil {
		log.Warn(e)
		return false
	}
	//必须配置超时
	if ttl <= 0 {
		ttl = redis_item_timeout
	}
	r, e := redis.String(conn.Do("Set", key, s, "EX", fmt.Sprintf("%d", ttl)))

	if e != nil {
		log.Warn(e)
	}
	if r == "OK" {
		//缓存本地
		if this.mem != nil {
			this.mem.Set(key, s, ttl)
		}
		//通告修改
		go this.publish(key, s, ttl, redis_sync_set)
		return true
	}
	log.Warn(r, ":", e)
	return false
}

func (this *RedisCache) Delete(key string) bool {
	conn := this.pool.Get()
	if conn == nil {
		return false
	}

	r, e := redis.Int(conn.Do("Del", key))
	if e != nil {
		log.Debug(e)
	}
	conn.Close()
	if r >= 0 {
		//缓存本地
		if this.mem != nil {
			this.mem.Delete(key)
		}
		//通告删除
		go this.publish(key, "", 0, redis_sync_del)
		return true
	}
	return true
}

func (this *RedisCache) Check(key string) bool {
	conn := this.pool.Get()
	if conn == nil {
		return false
	}
	defer conn.Close()

	o, e := conn.Do("Exists", key)

	if e != nil {
		log.Warn(e)
		return false
	}
	if o == int64(1) {
		return true
	}
	return false
}

func (this *RedisCache) CheckMem(key string) bool {
	return this.mem.Check(key)
}

func (this *RedisCache) SetTextSerialize(text TextSerialize) {
	this.text = text
}
