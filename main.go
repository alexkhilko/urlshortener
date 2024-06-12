package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
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
	for i := 0; i < 10; i++ {
		key := randString(10)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := redisClient.GetSet(ctx, key, url).Result()
		if err == redis.Nil {
			return key, nil
		} else if err != nil {
			return key, err
		}
	}
	return "", errors.New("failed to generate unique key after 10 attempts")
}

func handlePost(w http.ResponseWriter, req *http.Request) {
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

func handleGet(w http.ResponseWriter, req *http.Request) {
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

func handleDelete(w http.ResponseWriter, req *http.Request) {
	key := strings.Trim(req.URL.Path, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res, err := redisClient.Del(ctx, key).Result()
	if err != nil {
		fmt.Println("error on deleting key from redis", err)
		http.Error(w, "Internal error delete", http.StatusInternalServerError)
		return
	}
	if res == 0 {
		fmt.Println("key not found", key)
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		handlePost(w, req)
	case "GET":
		handleGet(w, req)
	case "DELETE":
		handleDelete(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	fmt.Println("Server is listening on :", port)
	http.HandleFunc("/", handler)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
