package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/alexkhilko/urlshortener/handler"
	"github.com/alexkhilko/urlshortener/repository"
	"github.com/redis/go-redis/v9"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("redis:%s", os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
		})
	repo := repository.NewApplicationRepository(redisClient)
	h := handler.NewAppHandler(repo)
	http.HandleFunc("/", h.Handle)
	fmt.Println("Server is listening on :", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
