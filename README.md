# Менеджер кеша в памяти на Golang



## Install

  go get github.com/milQA/mamorycache


## Inport

	import (
		memorycache "github.com/milQA/mamorycache"
	)

## Create cache

  cache := memorycache.New(Время хранения в кэше RAM, Интервал чистки, Время до удаления из кэша)

	cache := memorycache.New(5 * time.Minute, 2 * time.Minute, 10 * time.Minute)


## Как использовать

	Чтобы добавить запись в кэш необходимо использовать команду cache.Set(Ключ, Значение, Время до удаления из кэша, Время хранения в кэше RAM)

	cache.Set("Key", "Value", 10 * time.Minute, 5 * time.Minute)

  Если durationDelete = 0 и durationTransfer = 0 запись будет храниться в кэше до удаления командой cache.Delete().

  Чтобы получить значение из кэша необходимо использовать команду cache.Get("myKey")

	i := cache.Get("Key")
