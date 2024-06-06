package main

import (
	"crypto/md5"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"github.com/redis/go-redis/v9"
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

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

var redisClient * redis.Client

func init() {
    redisClient = redis.NewClient(&redis.Options{
        Addr:     "redis:6379",
        Password: "my-password", // no password set
        DB:       0,  // use default DB
    })
}

func handlePost(w http.ResponseWriter, req *http.Request) {
	var request ShortenURLRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		fmt.Println("error has occured", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	key := getMD5Hash(request.URL)[10:]
	e := redisClient.Set(context.Background(), key, request.URL, 0).Err()
	w.Header().Set("Content-Type", "application/json")
	if e != nil {
		fmt.Println("error on setting key in redis", e)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(NewShortenURLResponse(key, request.URL))
}

func handleGet(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	val, err := redisClient.Get(context.Background(), path).Result()
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

func handler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		handlePost(w, req)
	case "GET":
		handleGet(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8090", nil)
}
