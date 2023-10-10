package main

import (
	"context"
	"fmt"
	"time"

	"github.com/allegro/bigcache/v3"
)

type TileCache struct {
	cache *bigcache.BigCache
}

func (c *TileCache) Init() error {
	config := bigcache.DefaultConfig(10 * time.Minute)
	cache, err := bigcache.New(context.Background(), config)

	if err != nil {
		return err
	}

	c.cache = cache

	return nil
}

func (c *TileCache) formatKey(Z uint32, X, Y float64) string {
	return fmt.Sprintf("%d/%d/%d", Z, int64(X), int64(Y))
}

func (c *TileCache) Set(Z uint32, X, Y float64, img []byte) {
	key := c.formatKey(Z, X, Y)
	c.cache.Set(key, img)
}

func (c *TileCache) Get(Z uint32, X, Y float64) []byte {
	key := c.formatKey(Z, X, Y)
	entry, _ := c.cache.Get(key)

	return entry
}
