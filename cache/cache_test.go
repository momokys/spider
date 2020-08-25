package cache

import (
	"log"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	c := NewCache(5)

	go func() {
		i := 1
		for {
			time.Sleep(2*time.Second)
			c.Enter(i)
			i++
		}
	}()

	go func() {
		for {
			log.Println(c.Out())
		}
	}()

	time.Sleep(time.Minute)
}
