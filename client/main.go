package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error
		ok := errors.As(err, &netErr)
		if ok && netErr.Timeout() {
			log.Fatalf("Request to %s timed out", "http://localhost:8080/cotacao")
		}
		log.Fatal(err)
	}

	root := &struct {
		Bid string `json:"bid"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(root); err != nil {
		log.Fatal(err)
	}

	result := fmt.Sprintf("Dólar:%s", root.Bid)

	if err = os.WriteFile("cotacao.txt", []byte(result), 0644); err != nil {
		var netErr net.Error
		ok := errors.As(err, &netErr)
		if ok && netErr.Timeout() {
			log.Fatalf("Save file operation timed out")
		}
		log.Fatal(err)
	}

	log.Printf("File cotacao.txt saved with bid %s", root.Bid)
}
