package dcache

import (
	"fmt"
	"gopkg.in/redis.v3"
	"libs/encrypt"
	"net"
	"time"
)

type RedisClusterCache struct {
	cluster  *redis.ClusterClient
	cli      *redis.Client
	ip       string
	port     int
	pass     string
	db       int
	poolsize int
	name     string
	mem      *MemCache
	text     TextSerialize
	logger 	Logger
}

func newRedisClusterCache(ip string, port int, pass string, db int, poolsize int, text TextSerialize, logger Logger) *RedisClusterCache {
	s := new(RedisClusterCache)
	s.ip = ip
	s.pass = pass
	s.port = port
	s.db = db
	s.logger =logger
	s.name = encrypt.GetRandomString(16)
	s.cli = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", ip, port),
		Password: pass,
		DB:       int64(db),
		PoolSize: poolsize,
		Dialer: func() (net.Conn, error) {
			return net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
		},
	})
	s.cluster = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{fmt.Sprintf("%s:%d", ip, port)},
		Password: pass,
	})
	s.mem = newMemCache()
	go s.subscribe()
	return s
}

//监听数据修改事件
func (this *RedisClusterCache) subscribe() {
	var pub publisher
new:
	r, e := this.cli.Subscribe(redis_sync_chan)
	if e != nil {
		this.logger.Warn("redis sync failed:", e)
		time.Sleep(time.Second)
		goto new
	}
	for {
		v, e := r.ReceiveMessage()
		if e != nil {
			this.logger.Warn("Receive failed:", e)
			time.Sleep(time.Second)
			goto new
		}
		if e := this.text.Unmarshal([]byte(v.Payload), &pub); e == nil {
			if pub.From != this.name {
				//log.Debug(v.Channel, "[", this.name, "]:message", v.Payload)
				if pub.Act == redis_sync_set {
					this.mem.Set(pub.Key, pub.Val, pub.Ttl)
				} else if pub.Act == redis_sync_del {
					this.mem.Delete(pub.Key)
				}
			}
		} else {
			this.logger.Warn(e)
		}
	}
}

//消息推送
func (this *RedisClusterCache) publish(key, val string, ttl int, act int) {
	p := new(publisher)
	p.Key = key
	p.Val = val
	p.Ttl = ttl
	p.Act = act
	p.From = this.name
	//转为字符串
	s, e := this.text.Marshal(p)
	if e != nil {
		this.logger.Warn(e)
		return
	}

	r := this.cli.Publish(redis_sync_chan, s)
	if r.Err() != nil {
		this.logger.Warn("Publish error:", r.Err())
	}
}

func (this *RedisClusterCache) Get(key string, data interface{}) bool {
	var s string
	if !this.mem.Get(key, &s) {
		status := this.cluster.Get(key)
		if status.Err() != nil {
			this.logger.Debug("key:[", key, "]notfound->", status.Err())
			return false
		}
		s = status.Val()
		if ttl := this.cluster.TTL(key); ttl.Err() == nil {
			//设置本地内存
			if e := this.text.Unmarshal([]byte(s), data); e != nil {
				this.logger.Warn(e)
				return false
			}
			this.logger.Debug("load:", key, " ttl[", ttl.Val(), ":", int(ttl.Val()/time.Second), "] from redis:")
			//内存提前5秒超时
			return this.mem.Set(key, s, int(ttl.Val()/time.Second)-5)
		}

	}

	if e := this.text.Unmarshal([]byte(s), data); e != nil {
		this.logger.Warn(e)
		return false
	}
	return true
}

func (this *RedisClusterCache) Set(key string, data interface{}, ttl int) bool {
	//转为字符串
	s, e := this.text.Marshal(data)
	if e != nil {
		this.logger.Warn(e)
		return false
	}
	//必须配置超时
	if ttl <= 0 {
		ttl = redis_item_timeout
	}

	status := this.cluster.Set(key, s, time.Second*time.Duration(ttl))

	if status.Err() == nil {
		//缓存本地
		if this.mem != nil {
			this.mem.Set(key, s, ttl)
		}
		//通告修改
		go this.publish(key, s, ttl, redis_sync_set)
		return true
	}
	this.logger.Debug("Set key:", key, " err->", status.Err())
	return false
}

func (this *RedisClusterCache) Delete(key string) bool {
	this.mem.Delete(key)
	go this.publish(key, "", 0, redis_sync_del)
	if r := this.cluster.Del(key); r.Err() != nil {
		this.logger.Debug("delete key:", key, "->", r.Err())
		return false
	}
	return true
}

func (this *RedisClusterCache) Check(key string) bool {
	return this.cluster.Exists(key).Val()
}

func (this *RedisClusterCache) CheckMem(key string) bool {
	return this.mem.Check(key)
}

func (this *RedisClusterCache) SetTextSerialize(text TextSerialize) {
	this.text = text
}
