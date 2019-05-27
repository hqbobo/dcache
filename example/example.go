package main

import (
	"github.com/hqbobo/dcache"
	"fmt"
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
	dcache.GetCache().Set("aaaa", "ccccc", 100)
	if dcache.GetCache().Get("aaaa", &val) {
		fmt.Println(val)
	}
	select {}

}
