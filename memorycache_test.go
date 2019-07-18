package main

import (
	"fmt"

	memorycache "github.com/milQA/mamorycache"
)

func main() {
	cache := memorycache.New(5*time.Minute, 2*time.Minute, 10*time.Minute)
	cache.Set("Key", "Value", 10*time.Minute, 5*time.Minute)
	i := cache.Get("Key")
	fmt.Println(i)

}
