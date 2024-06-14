package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/alexkhilko/urlshortener/router"

)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}
	fmt.Println("Server is listening on :", port)
	http.HandleFunc("/", router.Router)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
