package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"
	"net/http"
	"time"
)

const (
	urlEconomia = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
)

type USDBRL struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	db, err := sql.Open("sqlite3", "database.sqlite3")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacao (bid TEXT, date TEXT DEFAULT CURRENT_TIMESTAMP)"); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	http.HandleFunc("GET /cotacao", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received")
		w.Header().Set("Content-Type", "application/json")

		ctxTimeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()

		client := &http.Client{}
		req, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, urlEconomia, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		resp, err := client.Do(req)
		if err != nil {
			var netErr net.Error
			ok := errors.As(err, &netErr)
			if ok && netErr.Timeout() {
				log.Printf("Request to %s timed out", urlEconomia)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		root := &struct {
			USDBRL USDBRL `json:"USDBRL"`
		}{}

		if err = json.NewDecoder(resp.Body).Decode(root); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		ctxTimeout, cancel = context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()

		if _, err = db.ExecContext(ctxTimeout, "INSERT INTO cotacao (bid) VALUES (?)", root.USDBRL.Bid); err != nil {
			var netErr net.Error
			ok := errors.As(err, &netErr)
			if ok && netErr.Timeout() {
				log.Printf("Database operation timed out")
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"bid": root.USDBRL.Bid})
	})

	log.Println("Server started at :8080")

	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
