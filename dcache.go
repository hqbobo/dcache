package dcache

import "encoding/json"
import "github.com/hqbobo/log"

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

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Notice(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Panic(v ...interface{})
}

type defaultLogger struct {
}

func (this *defaultLogger) Debug(v ...interface{})  { log.Debug(v...) }
func (this *defaultLogger) Info(v ...interface{})   { log.Info(v...) }
func (this *defaultLogger) Notice(v ...interface{}) { log.Notice(v...) }
func (this *defaultLogger) Warn(v ...interface{})   { log.Warn(v...) }
func (this *defaultLogger) Error(v ...interface{})  { log.Error(v...) }
func (this *defaultLogger) Panic(v ...interface{})  { log.Panic(v...) }
func (this *defaultLogger) Alert(v ...interface{})  { log.Alert(v...) }
func (this *defaultLogger) Fatal(v ...interface{})  { log.Fatal(v...) }

var cache Cache

type defaultText struct {
}

func (this *defaultText) Marshal(o interface{}) (string, error) {
	b, e := json.Marshal(o)
	return string(b), e
}

func (this *defaultText) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type Options struct {
	Ip          string        //redis server ip
	Port        int           //redis server port
	Pass        string        //redis server password
	Db          int           //redis server db
	PoolSize    int           //redis server clientpool size
	ClusterMode bool          //redis server running cluster mode?
	Serialize   TextSerialize //interface used to marshal object into text
	Logger      Logger        //interface used for log
}

func Init(option Options) {
	text := option.Serialize
	logger := option.Logger

	if text == nil {
		text = &defaultText{}
		log.InitLog(true ,log.AllLevels...)
		log.SetCode(true)
		log.SetPathFilter("github.com/hqbobo/")
	}

	if logger == nil {
		logger = &defaultLogger{}
	}

	if option.ClusterMode {
		cache = newRedisClusterCache(option.Ip, option.Port, option.Pass, option.Db, option.PoolSize, text, logger)
	} else {
		cache = newRedisCache(option.Ip, option.Port, option.Pass, option.Db, option.PoolSize, text, logger)
	}
}

func GetCache() Cache {
	return cache
}
