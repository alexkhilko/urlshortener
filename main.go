package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

// var cache map[string]string

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
		ShortURL: fmt.Sprintf("http://localhost:8905/%s", key),
		LongURL:  longURL,
	}
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func handlePost(w http.ResponseWriter, req *http.Request) {
	var request ShortenURLRequest
	err := json.NewDecoder(req.Body).Decode(&request)
	if err != nil {
		fmt.Println("error has occured", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	key := GetMD5Hash(request.URL)[10:]
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(NewShortenURLResponse(key, request.URL))
}

func handler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		handlePost(w, req)
	case "GET":
		fmt.Fprintf(w, "GET url")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8090", nil)
}
