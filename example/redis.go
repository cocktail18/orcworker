package main

import (
	"github.com/cocktail18/orcworker"
	"github.com/cocktail18/orcworker/downloader"
	"github.com/cocktail18/orcworker/storage"
	"log"
	"io/ioutil"
	"strconv"
)

func main() {
	download := downloader.NewSimpleDownloader()
	seed, _ := orcworker.NewSeed("https://blog.fly123.tk/echo.php?hello=1", "GET", nil, nil)
	st, err := storage.NewRedisStorage("127.0.0.1", 6379, "", 0)
	if err != nil {
		log.Fatalln(err)
	}
	//st := storage.NewMemoryStorage()
	scheduler, _ := orcworker.NewScheduler(st, download, -1 ,-1, nil)

	scheduler.AddProcessor(func(job *orcworker.Job) (seeds []*orcworker.Seed, err error) {
		body, _ := ioutil.ReadAll(job.Response.Body)
		log.Println(string(body))
		for i := 0; i < 10; i++ {
			seed, _ := orcworker.NewSeed("https://blog.fly123.tk/echo.php?hello="+strconv.Itoa(i), "GET", nil ,nil)
			seeds = append(seeds, seed)
		}
		return
	})
	log.Fatalln(scheduler.Start(seed))
}
