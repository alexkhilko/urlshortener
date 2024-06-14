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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println("checking url in redis", url)
	val, err := redisClient.Get(ctx, url).Result()
	if err == nil {
		return val, nil
	}

	for i := 0; i < 5; i++ {
		key := randString(10)
		_, err := redisClient.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			val, err := pipe.Exists(ctx, key).Result()
			fmt.Println("checking value exists", val, err)
			if err != nil {
				return err
			}
			if val > 0 {
				return errors.New("value exists")
			}
			e1 := pipe.Set(ctx, key, url, 0).Err()
			fmt.Println("setting key", key, url)

			if e1 != nil {
				fmt.Println("error setting key", e1)
				return e1
			}
			fmt.Println("setting url", url, key)
			e2 := pipe.Set(ctx, url, key, 0).Err()
			if e2 != nil {
				fmt.Println("error setting key", e1)
				return e2
			}
			return nil
		})
		if err == nil {
			return key, nil
		}
	}
	return "", errors.New("failed to generate unique key after 5 attempts")
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
	// TODO: Make transactional delete
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
