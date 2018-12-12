package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/cocktail18/orcworker"
	"github.com/cocktail18/orcworker/downloader"
	"github.com/cocktail18/orcworker/storage"
	"gopkg.in/bsm/ratelimit.v2"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	download := downloader.NewSimpleDownloader()
	seed, _ := orcworker.NewSeed("https://www.mzitu.com/", "GET", nil, nil)
	st := storage.NewMemoryStorage()
	rate := ratelimit.New(100, time.Second)
	scheduler, _ := orcworker.NewScheduler(st, download, -1, 100, rate)
	scheduler.AddProcessor(func(job *orcworker.Job) (seeds []*orcworker.Seed, err error) {
		doc, err := goquery.NewDocumentFromReader(job.Response.Body)
		job.Result["doc"] = doc
		if doc.Find("#pins") == nil {
			return
		}
		doc.Find("#pins li").Each(func(i int, selection *goquery.Selection) {
			a := selection.Find("a").Eq(0)
			href, b := a.Attr("href")
			if b {
				seed, _ := orcworker.NewSeed(href, "GET", nil, nil)
				seeds = append(seeds, seed)
			}
		})
		doc.Find(".pagination .nav-links a.page-numbers").Each(func(i int, selection *goquery.Selection) {
			if href, b := selection.Attr("href"); b {
				seed, _ := orcworker.NewSeed(href, "GET", nil, nil)
				seeds = append(seeds, seed)
			}
		})
		return
	})

	scheduler.AddProcessor(func(job *orcworker.Job) (seeds []*orcworker.Seed, err error) {
		doc, _ := job.Result["doc"].(*goquery.Document)
		title := doc.Find(".main-title")
		if title == nil {
			return
		}
		doc.Find(".main-image p a").Each(func(i int, selection *goquery.Selection) {
			href, b := selection.Attr("href")
			if b {
				seed, _ := orcworker.NewSeed(href, "GET", nil, nil)
				seeds = append(seeds, seed)
			}
			src, b := selection.Find("img").Attr("src")
			if b {
				dst := "output/" + title.Text() + ".jpg"
				downloadImage(src, dst)
			}
		})
		doc.Find(".pagenavi a").Each(func(i int, selection *goquery.Selection) {
			if href, b := selection.Attr("href"); b {
				seed, _ := orcworker.NewSeed(href, "GET", nil, nil)
				seeds = append(seeds, seed)
			}
		})
		for i := 0; i < len(seeds); i++ {
			if strings.HasPrefix(seeds[i].URL, "/") {
				seeds[i].URL = "https://www.mzitu.com" + seeds[i].URL
			}
		}
		return
	})

	go func() {
		for {
			<-time.After(time.Second)
			scheduler.Stat()
		}
	}()
	log.Fatalln(scheduler.Start(seed))
}

func downloadImage(src, dst string) error {
	client := http.Client{}
	req, err := http.NewRequest("GET", src, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Referer", "https://www.mzitu.com/")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.102 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, body, 0755)
}
