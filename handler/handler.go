package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/http"
	"os"
	"strings"
	"time"
)

type ShortenURLRequest struct {
	URL string `json:"url"`
}

type ShortenURLResponse struct {
	Key      string `json:"key"`
	ShortURL string `json:"short_url"`
	LongURL  string `json:"long_url"`
}

func NewShortenURLResponse(key string, longURL string) ShortenURLResponse {
	return ShortenURLResponse{
		Key:      key,
		ShortURL: fmt.Sprintf("http://localhost:8090/%s", key),
		LongURL:  longURL,
	}
}

var redisClient *redis.Client

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("redis:%s", os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func generateShortURL(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	val, err := redisClient.Get(ctx, url).Result()
	if err == nil {
		return val, nil
	}
	key := randString(10)
	e1 := redisClient.Set(ctx, key, url, 0).Err()
	if e1 != nil {
		fmt.Println("error setting key", e1)
		return "", e1
	}
	e2 := redisClient.Set(ctx, url, key, 0).Err()
	if e2 != nil {
		fmt.Println("error setting key value", e2)
		return "", e2
	}
	return key, nil
}

func Post(w http.ResponseWriter, req *http.Request) {
	var request ShortenURLRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		fmt.Println("error has occured", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	key, e := generateShortURL(request.URL)
	if e != nil {
		fmt.Println("error on setting key in redis", e)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(NewShortenURLResponse(key, request.URL))
}

func Get(w http.ResponseWriter, req *http.Request) {
	key := strings.Trim(req.URL.Path, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("error on getting key from redis", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, val, http.StatusMovedPermanently)
}

func Delete(w http.ResponseWriter, req *http.Request) {
	key := strings.Trim(req.URL.Path, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	val, err1 := redisClient.GetDel(ctx, key).Result()
	if err1 == redis.Nil {
		fmt.Println("key not found", key)
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err1 != nil {
		fmt.Println("error on deleting key from redis", err1)
		http.Error(w, "Internal error delete", http.StatusInternalServerError)
		return
	}
	err2 := redisClient.Del(ctx, val).Err()
	if err2 != nil {
		fmt.Println("long url was not deleted", key)
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
