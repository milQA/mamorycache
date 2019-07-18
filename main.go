package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

//==============================

type Cache struct {
	sync.RWMutex
	items             map[string]Item
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	expirationTime    time.Duration
	itemsSecondCache  map[string]int64
}

type Item struct {
	Value                        interface{} `json:"value"`
	Created                      time.Time   `json:"created"`
	ExpirationDeleteTime         int64       `json:"expiration"`
	TransferCacheSecondLevelTime int64       `json:"transfer"`
}

//===============================

func main() {
	cache := New(5*time.Minute, 2*time.Minute, 10*time.Minute)
	cache.Set("Key", "Value", 10*time.Minute, 5*time.Minute)
	cache.transferItems([]string{"Key"})
	i, _ := cache.Get("Key")
	fmt.Print(i)
}

func New(defaultExpiration, cleanupInterval, expirationTime time.Duration) *Cache {

	items := make(map[string]Item)
	itemsSecondCache := make(map[string]int64)

	cache := Cache{
		items:             items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		expirationTime:    expirationTime,
		itemsSecondCache:  itemsSecondCache,
	}

	//лог: "make structure"

	if cleanupInterval > 0 {
		cache.StartGC()
	}

	return &cache
}

func (c *Cache) Set(key string, value interface{}, durationDelete, durationTransfer time.Duration) {

	var expirationDeleteTime, transferCacheSecondLevelTime int64

	if durationDelete == 0 {
		durationDelete = c.expirationTime
	}

	if durationTransfer == 0 {
		durationTransfer = c.defaultExpiration
	}

	if durationDelete > 0 {
		expirationDeleteTime = time.Now().Add(durationDelete).UnixNano()
	}

	if durationTransfer > 0 {
		transferCacheSecondLevelTime = time.Now().Add(durationTransfer).UnixNano()
	}

	c.Lock()

	c.items[key] = Item{
		Value:                        value,
		Created:                      time.Now(),
		ExpirationDeleteTime:         expirationDeleteTime,
		TransferCacheSecondLevelTime: transferCacheSecondLevelTime,
	}

	// лог: "structure "key" is added in RAM"

	c.Unlock()

}

func (c *Cache) Get(key string) (interface{}, bool) {

	c.RLock()

	defer c.RUnlock()

	item, found := c.items[key] //поиск с RAM

	if !found {
		item, found = c.GetSecondCache(key) // поиск с HDD
		if !found {
			return nil, false
		}
	}

	//Ниже проверка на жизнь найденной записи
	if item.ExpirationDeleteTime > 0 {
		if time.Now().UnixNano() > item.ExpirationDeleteTime {
			return nil, false
		}
	}

	return item.Value, true
}

func (c *Cache) GetSecondCache(key string) (Item, bool) {

	c.RLock()

	defer c.RUnlock()

	_, found := c.itemsSecondCache[key]

	if !found {
		return Item{}, false
	}
	f, err := os.Open(key)
	if err != nil {
		fmt.Println(err)
	}
	fi, _ := f.Stat()
	by := make([]byte, fi.Size())
	_, err = f.Read(by)
	if err != nil {
		fmt.Println(err)
	}
	var item Item
	err = json.Unmarshal(by, item)
	if err != nil {
		fmt.Println(err)
	}
	// Нужно сделать чтение с файла под именем key сруктуры Item в переменную item

	c.items[key] = item

	delete(c.itemsSecondCache, key)

	//лог: "structure "key" moveed is HDD in RAM"
	return item, true
}

func (c *Cache) Delete(key string) error {

	c.Lock()

	defer c.Unlock()

	if _, found := c.items[key]; !found {
		found := c.deleteSecondCache(key)
		if !found {
			return errors.New("Key not found")
		} else {
			return nil
		}
	}

	delete(c.items, key)
	//лог: "structure "key" delete"

	return nil
}

func (c *Cache) StartGC() {
	go c.GC()
}

func (c *Cache) GC() {

	for {

		<-time.After(c.cleanupInterval)

		if c.items == nil {
			return
		}

		if keys := c.transferKeys(); len(keys) != 0 {
			c.transferItems(keys)
		}

		if keys := c.expiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}
	}

	c.CacheStatus()
}

func (c *Cache) CacheStatus() {

	//лог: "number of structures in RAM: len(c.items)"
	//лог: "number of structures in HDD: len(c.itemsSecondCache)"

}

func (c *Cache) expiredKeys() (keys []string) {

	c.RLock()

	defer c.RUnlock()

	for k, i := range c.items {
		if time.Now().UnixNano() > i.ExpirationDeleteTime && i.ExpirationDeleteTime > 0 {
			keys = append(keys, k)
		}
	}

	for k, i := range c.itemsSecondCache {
		if time.Now().UnixNano() > i && i > 0 {
			keys = append(keys, k)
		}
	}

	return
}

func (c *Cache) transferKeys() (keys []string) {

	c.RLock()

	defer c.RUnlock()

	for k, i := range c.items {
		if time.Now().UnixNano() > i.TransferCacheSecondLevelTime && i.TransferCacheSecondLevelTime > 0 {
			keys = append(keys, k)
		}
	}

	return
}

func (c *Cache) clearItems(keys []string) {

	c.Lock()

	for _, k := range keys {
		err := os.Remove(k)
		if err != nil {
			delete(c.items, k)
		} else {
			delete(c.itemsSecondCache, k)
		}
	}

	//лог: "structure "key" delete"

	c.Unlock()
}

func (c *Cache) transferItems(keys []string) {

	c.Lock()
	for _, key := range keys {
		c.itemsSecondCache[key] = c.items[key].ExpirationDeleteTime
		f, err := os.Create(key)
		if err != nil {
			fmt.Println(err)
		}
		by, err := json.Marshal(c.items[key])
		if err != nil {
			fmt.Println(err)
		}
		_, err = f.Write(by)
		if err != nil {
			fmt.Println(err)
		}
		f.Close()

		delete(c.items, key)

		//лог: "structure "key" moveed is RAM in HDD"

	}

	c.Unlock()
}

func (c *Cache) deleteSecondCache(key string) bool {

	c.Lock()

	defer c.Unlock()

	err := os.Remove(key)
	if err != nil {
		return false
	}

	delete(c.itemsSecondCache, key)
	//лог: "structure "key" delete"
	return true
}
