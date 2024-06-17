package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"github.com/alexkhilko/urlshortener/repository"
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


func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

type AppHandler struct {
	repo repository.Repository
}

func NewAppHandler(repo repository.Repository) *AppHandler {
	return &AppHandler{repo: repo}
}

func generateShortURL(repo repository.Repository,  url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	val, err := repo.Find(ctx, url)
	if err == nil && val != "" {
		return val, nil
	}
	key := randString(10)
	e1 := repo.Set(ctx, key, url)
	if e1 != nil {
		fmt.Println("error setting key", e1)
		return "", e1
	}
	e2 := repo.Set(ctx, url, key)
	if e2 != nil {
		fmt.Println("error setting key value", e2)
		return "", e2
	}
	return key, nil
}

func (a *AppHandler) Post(w http.ResponseWriter, req *http.Request) {
	var request ShortenURLRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		fmt.Println("error has occured", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	key, e := generateShortURL(a.repo, request.URL)
	if e != nil {
		fmt.Println("error on setting key", e)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(NewShortenURLResponse(key, request.URL))
}

func (a *AppHandler) Get(w http.ResponseWriter, req *http.Request) {
	key := strings.Trim(req.URL.Path, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	val, err := a.repo.Find(ctx, key)
	if err != nil {
		fmt.Println("error on getting key from repo", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	} else if val == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	http.Redirect(w, req, val, http.StatusMovedPermanently)
}

func (a *AppHandler) Delete(w http.ResponseWriter, req *http.Request) {
	key := strings.Trim(req.URL.Path, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	val, err1 := a.repo.GetDel(ctx, key)
	if err1 != nil {
		fmt.Println("error on deleting key", err1)
		http.Error(w, "Internal error delete", http.StatusInternalServerError)
		return
	} else if val == "" {
		fmt.Println("key not found", key)
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} 
	_, err2 := a.repo.GetDel(ctx, val)
	if err2 != nil {
		fmt.Println("long url was not deleted", key)
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}
