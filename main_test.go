package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"github.com/alexkhilko/urlshortener/handler"
	"github.com/alexkhilko/urlshortener/repository"
	"time"
)

var  (
	port string
	testUrl string = "http://test.com"
	testKey string = "foo"
	client *http.Client
)


func TestMain(m *testing.M) {
	port = "9989"
	client = &http.Client{
        Timeout: 5 * time.Second, 
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
            // Prevent the client from following redirects
            return http.ErrUseLastResponse
        },
    }
	go func() {
		repo := repository.NewTestRepository(map[string]string{testKey: testUrl})
		h := handler.NewAppHandler(repo)
		http.HandleFunc("/", h.Handle)
		http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	}()
	exitVal := m.Run()
	// Clean up
	os.Exit(exitVal)
}


func TestMissingItem(t *testing.T) {
	resp, err := client.Get(fmt.Sprintf("http://localhost:%s/bar/", port))
	if err != nil {
		t.Error("error on calling api", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Error("unexpected status code", resp.StatusCode)
	} 
}


func TestRedirect(t *testing.T) {
	url := fmt.Sprintf("http://localhost:%s/%s", port, testKey)
	resp, err := client.Get(url)
	if err != nil {
		t.Error("error on calling api", err)
	}
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Error("unexpected status code", resp.StatusCode)
	}
	location, err := resp.Location()
	if err != nil {
        t.Error("Error getting redirect location", err)
    }
	if testUrl != location.String(){
        t.Error("Redirect URL mismatch (-want +got):", testUrl, location.String())
    }
}

func TestShortenURL(t *testing.T) {
	url := fmt.Sprintf("http://localhost:%s/", port)
	longURL := "http://example.com"
	requestBody, err := json.Marshal(handler.ShortenURLRequest{URL: longURL})
	if err != nil {
		t.Error("Error marshalling request body", err)
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Error("Error making POST request:", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected status code", resp.StatusCode)
	}
}