package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"github.com/google/go-cmp/cmp"
)

func TestMissingItem(t *testing.T) {
	port := os.Getenv("PORT")
	resp, err := http.Get(fmt.Sprintf("http://web:%s/foo", port))
	if err != nil {
		t.Error("error on calling api", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		b, _ := io.ReadAll(resp.Body)
		t.Error("unexpected status code", resp.StatusCode, string(b))
	} 
}

func TestShortenURL(t *testing.T) {
	port := os.Getenv("PORT")
	url := fmt.Sprintf("http://web:%s/", port)
	longURL := "http://example.com"
	requestBody, err := json.Marshal(ShortenURLRequest{URL: longURL})
	if err != nil {
		t.Error("Error marshalling request body", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Error("Error making POST request:", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected status code", resp.StatusCode)
	}
	var shortenResponse ShortenURLResponse
    if err := json.NewDecoder(resp.Body).Decode(&shortenResponse); err != nil {
		t.Error("Error decoding response body:", err)
    }
	expKey := getMD5Hash(longURL)[10:]
	expected := ShortenURLResponse{
		Key: expKey,
		ShortURL: fmt.Sprintf("http://localhost:%s/%s", port, expKey),
		LongURL: longURL,
	}
	if diff := cmp.Diff(shortenResponse, expected); diff != "" {
		t.Error(diff)
	}
}