package main

import (
	"fmt"
	"net/http"
	"strings"
	"service/memorycache"
	"github.com/gorilla/mux"
	"encoding/json"
	"log"
	"time"
)

type Cache struct {
	defaultExpiration time.Duration `json: transfer`
	cleanupInterval   time.Duration `json: interval`
	expirationTime    time.Duration `json: expiration`
}

type Item struct {
	key   string      `json: key`
	value interface{} `json:"value"`
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handler)
	s := r.PathPrefix("/cache").Subrouter()
	s.HandleFunc("/Make", handlerPostMakeCache).Methods("POST")   			// вызвать New(JSON)
	s.HandleFunc("/Status", handlerGetCacheStatus).Methods("GET")				// узнать статус кэша
	s.HandleFunc("/{key}", handlerGetCacheValue).Methods("GET")					// вызвать Get(key)
	s.HandleFunc("/Add", handlerPostCacheValue).Methods("POST")					// вызвать Set(JSON)
	s.HandleFunc("/{key}", handlerDeleteCacheValue).Methods("DELETE")		// вызвать Delete(key)

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8000", rout))
}

func handler(w http.ResponseWriter, r *http.Request) {
	return
}

func handlerPostMakeCache(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var item Cache
	err := json.NewDecoder(r.Body).Decode(&item)


	var cache memorycache.Cache
	cache := memorycache.New(item.defaultExpiration, item.cleanupInterval, item.expirationTime)

	json.NewEncoder(w).Encode("Все найз")

	// Нужно сделать чтение с JSON. JSON из 3-х полей

	return
}

func handlerGetCacheStatus(w http.ResponseWriter, r *http.Request) {

	cache.CacheStatus()

	//необходимо сделать вывод статуса кэша в JSON

	return
}

func handlerGetCacheValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	i, _ := cache.Get(vars["key"])
	item := Item{
		key:   vars["key"],
		value: i,
	}

	json.NewEncoder(w).Encode(&item)

}

func handlerPostCacheValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Нужно сделать чтение с JSON. JSON из 4-х полей

	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)

	if err != nil {
		json.NewEncoder(w).Encode("400 Bad Request")
		return
	}

	cache.Set(item.key, item.value, 0, 0)

	json.NewEncoder(w).Encode("Все найз")

	return
}

func handlerDeleteCacheValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)

	err := cache.Delete(vars["key"])

	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}

	return
}
