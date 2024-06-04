package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type DB	struct {
	shortToLong map[string]string
	longToShort map[string]string
}

func (db DB) add(short string, long string) {
	db.shortToLong[short] = long
	db.longToShort[long] = short
}

var db DB

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

func getMD5Hash(text string) string {
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
	key := getMD5Hash(request.URL)[10:]
	w.Header().Set("Content-Type", "application/json")
	db.add(key, request.URL)
	json.NewEncoder(w).Encode(NewShortenURLResponse(key, request.URL))
}

func handleGet(w http.ResponseWriter, req *http.Request) {
	path := strings.Trim(req.URL.Path, "/")
	fmt.Println(path)
	fmt.Println(db.shortToLong)
	url, ok := db.shortToLong[path]
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	http.Redirect(w, req, url, http.StatusMovedPermanently)
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
	db = DB{shortToLong: map[string]string{}, longToShort: map[string]string{}}
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8090", nil)
}
