package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

//O client.go deverá realizar uma requisição HTTP no server.go solicitando a cotação do dólar.
//
//O server.go deverá consumir a API contendo o câmbio de Dólar e Real no endereço: https://economia.awesomeapi.com.br/json/last/USD-BRL e em seguida deverá retornar no formato JSON o resultado para o cliente.
//
//Usando o package "context", o server.go deverá registrar no banco de dados SQLite cada cotação recebida, sendo que o timeout máximo para chamar a API de cotação do dólar deverá ser de 200ms e o timeout máximo para conseguir persistir os dados no banco deverá ser de 10ms.
//
//O client.go precisará receber do server.go apenas o valor atual do câmbio (campo "bid" do JSON). Utilizando o package "context", o client.go terá um timeout máximo de 300ms para receber o resultado do server.go.
//
//Os 3 contextos deverão retornar erro nos logs caso o tempo de execução seja insuficiente.
//
//O client.go terá que salvar a cotação atual em um arquivo "cotacao.txt" no formato: Dólar: {valor}
//
//O endpoint necessário gerado pelo server.go para este desafio será: /cotacao e a porta a ser utilizada pelo servidor HTTP será a 8080.
//
//Ao finalizar, envie o link do repositório para correção.

const (
	urlEconomia = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
)

func main() {
	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if _, err = db.Exec("CREATE TABLE IF NOT EXISTS cotacao (bid TEXT)"); err != nil {
		log.Fatal(err)
	}

	r := http.NewServeMux()

	r.HandleFunc("GET /cotacao", func(w http.ResponseWriter, r *http.Request) {
		data, err := getRequest(urlEconomia)
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

		if _, err = db.Exec("INSERT INTO cotacao (bid) VALUES (?)", root.USDBRL.Bid); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"bid": root.USDBRL.Bid})
	})

	if err = http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

func getRequest(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 200 * time.Millisecond}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
