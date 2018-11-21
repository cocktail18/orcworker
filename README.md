Orcworker 是一款纯Go实现的、轻量级、插件化的 爬虫框架


## Go版本要求

≥Go1.10

## example
```
package main

import (
	"github.com/cocktail18/orcworker"
	"log"
	"io/ioutil"
)

func main() {
	fetcher := orcworker.NewSimpleFetcher()
	seed, _ := orcworker.NewSeed("https://blog.fly123.tk/json.php", "GET", nil, nil, 1)
	seedsBank := orcworker.NewMemorySeedsBank()
	scheduler, _ := orcworker.NewScheduler(seedsBank, fetcher, -1, nil)

	scheduler.AddProcessor(func(job *orcworker.Job) (seeds []*orcworker.Seed, err error) {
		body, _ := ioutil.ReadAll(job.Response.Body)
		log.Println(string(body))
		return
	})
	log.Fatalln(scheduler.Start(seed))
}

```

### 更多例子看 example 目录
