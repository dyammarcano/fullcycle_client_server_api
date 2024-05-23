package components

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	urlEconomia = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
)

type Root struct {
	USDBRL USDBRL `json:"USDBRL"`
}

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

func getRequest(ctx context.Context, url string) (*http.Response, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error
		ok := errors.As(err, &netErr)
		if ok && netErr.Timeout() {
			log.Printf("Request to %s timed out", url)
		}
		return nil, err
	}
	return resp, nil
}

func Client(ctx context.Context, port string) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	data, err := getRequest(ctxTimeout, fmt.Sprintf("http://localhost:%s/cotacao", port))
	if err != nil {
		log.Fatal(err)
	}

	root := &Root{}
	if err = json.NewDecoder(data.Body).Decode(root); err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("cotacao.txt", []byte(root.USDBRL.Bid), 0644); err != nil {
		var netErr net.Error
		ok := errors.As(err, &netErr)
		if ok && netErr.Timeout() {
			log.Printf("Save file operation timed out")
		}
		log.Fatal(err)
	}

	log.Printf("File cotacao.txt saved with bid %s", root.USDBRL.Bid)
}

func Server(ctx context.Context, port, filePath string) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacao (bid TEXT, date TEXT DEFAULT CURRENT_TIMESTAMP)"); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("GET /cotacao", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received")
		data, err := getRequest(ctx, urlEconomia)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		root := &Root{}
		if err = json.NewDecoder(data.Body).Decode(root); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
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

	log.Printf("Server started at :%s", port)

	if err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
}
