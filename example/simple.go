package main

import (
	"github.com/cocktail18/orcworker"
	"github.com/cocktail18/orcworker/downloader"
	"github.com/cocktail18/orcworker/storage"
	"log"
	"io/ioutil"
)

func main() {
	download := downloader.NewSimpleDownloader()
	seed, _ := orcworker.NewSeed("https://blog.fly123.tk/json.php", "GET", nil, nil)
	st := storage.NewMemoryStorage()
	scheduler, _ := orcworker.NewScheduler(st, download, -1 ,-1, nil)

	scheduler.AddProcessor(func(job *orcworker.Job) (seeds []*orcworker.Seed, err error) {
		body, _ := ioutil.ReadAll(job.Response.Body)
		log.Println(string(body))
		return
	})
	log.Fatalln(scheduler.Start(seed))
}
