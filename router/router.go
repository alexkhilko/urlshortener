package router

import (
	"net/http"
	"github.com/alexkhilko/urlshortener/handler"
	"github.com/alexkhilko/urlshortener/repository"
	"github.com/redis/go-redis/v9"
	"os"
	"fmt"

)

func Router(w http.ResponseWriter, req *http.Request) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("redis:%s", os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	repo := repository.NewApplicationRepository(redisClient)
	h := handler.NewAppHandler(repo)
	switch req.Method {
	case "POST":
		h.Post(w, req)
	case "GET":
		h.Get(w, req)
	case "DELETE":
		h.Delete(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
