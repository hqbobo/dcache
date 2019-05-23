
# Dcache  
  
## Description  
  
>distribute cache implement with redis. local memory is also used which can speed up get operation.<br>
main function as below

1. support redis and redis cluster mode  
2. support local cache. all set operation will remian a copy of data in memory(with timeout).
3. All dcache client will get a copy of data after set opertion and save in their own memory(with timeout).
4. All data has been marshelled into json, feel free to store struct stuff
5. Set your own logger and TextSerialize


##Detail
> Option
```$xslt
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
```

> TextSerialize

you can design your own TextSerialize if speed is concerned
```$xslt
type TextSerialize interface {
	Marshal(o interface{}) (string, error)
	Unmarshal(data []byte, v interface{}) error
}
```

> Logger

fit to the log system that you are familiar with
```$xslt
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Notice(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Panic(v ...interface{})
}
```

## Example  
> [Source Code](https://github.com/hqbobo/dcache/example)  see in the example directory
```
package main

import (
	"github.com/hqbobo/dcache"
)

func main() {
	dcache.Init(dcache.Options{
		Ip:          "127.0.0.1",
		Port:        6379,
		Pass:        "",
		Db:          1,
		PoolSize:    10,
		ClusterMode: false,
	})
	var val string
	dcache.GetCache().Get("aaaa", &val)
	dcache.GetCache().Set("aaaa", "bbbb", 100)
	dcache.GetCache().Get("aaaa", &val)
	select {}

}

```