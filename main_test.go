package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"github.com/redis/go-redis/v9"
	"github.com/google/go-cmp/cmp"
	"context"
	"time"
)

var  (
	port string
	testUrl string = "http://test.com"
	testKey string = "foo"
	client *http.Client
)


func TestMain(m *testing.M) {
	port = os.Getenv("PORT")
	redisClient = redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("redis:%s", os.Getenv("REDIS_PORT")),
        Password: os.Getenv("REDIS_PASSWORD"),
        DB:       0,
    })
	client = &http.Client{
        Timeout: 5 * time.Second, 
    }
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()
	e := redisClient.Set(ctx, testKey, testUrl, 0).Err()
	if e != nil {
		panic("Could not set up test data in Redis: " + e.Error())
	}
	// Ensure Redis is reachable
	if err := redisClient.Ping(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to Redis: %v\n", err)
		os.Exit(1)
	}
	exitVal := m.Run()
	// Clean up
	if err := redisClient.FlushDB(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not clean up test data in Redis: %v\n", err)
	}
	os.Exit(exitVal)
}


func TestMissingItem(t *testing.T) {
	resp, err := client.Get(fmt.Sprintf("http://web:%s/foo", port))
	if err != nil {
		t.Error("error on calling api", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Error("unexpected status code", resp.StatusCode)
	} 
}


func TestRedirect(t *testing.T) {
	url := fmt.Sprintf("http://web:%s/%s", port, testKey)
	fmt.Println("URL ", url)
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
	url := fmt.Sprintf("http://web:%s/", port)
	longURL := "http://example.com"
	requestBody, err := json.Marshal(ShortenURLRequest{URL: longURL})
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