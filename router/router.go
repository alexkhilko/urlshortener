package router

import (
	"net/http"
	"github.com/alexkhilko/urlshortener/handler"
)

func Router(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		handler.Post(w, req)
	case "GET":
		handler.Get(w, req)
	case "DELETE":
		handler.Delete(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
