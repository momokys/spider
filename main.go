package main

import (
	"flag"
	"log"
	"momoky.cn/spider/catch"
	"sync"

)

func main()  {
	var title string
	var path string

	flag.StringVar(&title, "t", "风景", "主题")
	flag.StringVar(&path, "p", "./", "下载路径")
	flag.Parse()

	log.Println(title,path)

	c := catch.NewCatch("https://rt.huashi6.com/front/works/search",
		"https://img2.huashi6.com/", title, 0,path,
		`(?:(?:"path")|(?:"coverImageUrl")):"(.+?)"`)
	var wait sync.WaitGroup
	isExit := false
	wait.Add(1)
	go func() {
		for {
			if err := c.CatchUrl(); err != nil {
				log.Println(err)
				isExit = true
				return
			}
		}
	}()

	for i := 0;i < 10;i++ {
		wait.Add(1)
		go func() {
			for !isExit {
				c.LoadImage()
			}
			wait.Done()
		}()
	}

	wait.Add(1)
	go func() {
		for !isExit {
			c.DownLoadImage()
		}
		wait.Done()
	}()

	wait.Wait()
}
