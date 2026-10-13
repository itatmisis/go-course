package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func doClient() {
	client := &http.Client{Timeout: 5 * time.Second}
	u, err := url.Parse("https://httpbin.org/get")
	if err != nil {
		log.Fatal(err)
	}

	q := u.Query()
	q.Set("language", "Go") // https://httpbin.org/get?language=Go
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		fmt.Println("client error")
		return
	}

	if resp.StatusCode >= 500 {
		fmt.Println("server error")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Status:", resp.Status)
	fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))
	fmt.Println(string(body))
}
