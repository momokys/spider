package catch

import (
	"fmt"
	"io/ioutil"
	"log"
	"momoky.cn/spider/cache"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

const (
	HOST = `https://rt.huashi6.com/front/works/search`
	REGEXP = `(?:(?:"path")|(?:"faceUrl")|(?:"coverImageUrl")):"(.+?)"`
	PREFIX = `https://img2.huashi6.com/`
)

type Catch struct {
	host string
	prefix string
	title string
	index int
	order int
	path string
	mutex sync.Mutex
	reg *regexp.Regexp
	nameReg *regexp.Regexp
	urlCache *cache.Cache
	imgCache *cache.Cache
}

type Image struct {
	imageBytes []byte
	name string
}


func NewCatch(host string, prefix string, title string, index int, path string, r string) *Catch {
	reg := regexp.MustCompile(r)
	nameReg := regexp.MustCompile(`^.*/((?:.*?\.jpg)|(?:.*?\.png)|(?:.*?\.jpeg))$`)
	urlCache := cache.NewCache(64)
	imgCache := cache.NewCache(64)
	catch := &Catch{
		host,
		prefix,
		title,
		index,
		0,
		path,
		sync.Mutex{},
		reg,
		nameReg,
		urlCache,
		imgCache,
	}

	return catch
}

// 获取时间戳
func (c *Catch) TimeStamp() string  {
	return fmt.Sprintf("%v", time.Now().UnixNano()/1000000)
}

// 获取url
func (c *Catch) ParseUrl() string {
	values := url.Values{}
	values.Set("_ts_", c.TimeStamp())
	values.Set("index", strconv.Itoa(c.index))
	values.Set("title", c.title)

	return c.host + "?" + values.Encode()
}

// 获取 url 后缀用于判断图片文件格式
func (c *Catch) Name(url string) string {
	match := c.nameReg.FindStringSubmatch(url)
	if len(match) < 2 {
		log.Println(url, match)
		return ""
	}
	return match[1]
}

func (c *Catch) getOrder() (order int)  {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.order++
	order = c.order
	return order
}

// 从 json 中匹配图片链接
func (c *Catch) Match(json *string) {
	matchSet := c.reg.FindAllStringSubmatch(*json, -1)
	for _, match := range matchSet {

		c.urlCache.Enter(c.prefix + match[1])
	}
}

// 从 host 指向的 api 中抓取图片链接
func (c *Catch) catchUrl(host string) error {
	resp, err := http.Get(host)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	jsonByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	jsonStr := string(jsonByte)
	c.Match(&jsonStr)
	return nil
}

func (c *Catch) CatchUrl() error {
	c.index++
	err := c.catchUrl(c.ParseUrl())
	return err
}

// 从 url 中加载图片到内存
func (c *Catch) loadImage(url string) *Image {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	img := &Image{}
	img.imageBytes = bytes
	img.name = c.Name(url)

	return img
}

func (c *Catch) LoadImage() {
	u, ok := c.urlCache.Out().(string)
	if ok {
		image := c.loadImage(u)
		c.imgCache.Enter(image)
		log.Println(u,"load finish!")
	}
}

// 将图片写入本地存储
func (c *Catch) downLoadImage(imgBytes []byte, fileName string) error {
	err := ioutil.WriteFile(fileName, imgBytes, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}


func (c *Catch) DownLoadImage() {
	img, ok := c.imgCache.Out().(*Image)
	if ok {
		err := c.downLoadImage(img.imageBytes, c.path+"/"+img.name)
		if err != nil {
			return
		}
		log.Println(c.path+"/"+img.name, "download finish!")
	}
}


