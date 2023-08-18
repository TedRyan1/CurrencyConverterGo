package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	apiURL        = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	databaseTimeout = 10 * time.Millisecond
	serverTimeout    = 200 * time.Millisecond
	serverAddress = ":8080"
)

type QuoteResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", handleCotacaoRequest)
	http.ListenAndServe(serverAddress, nil)
}

func handleCotacaoRequest(w http.ResponseWriter, r *http.Request) {
	bid, err := fetchBidFromAPI()
	if err != nil {
		http.Error(w, "Error fetching quote", http.StatusInternalServerError)
		fmt.Println("API error:", err)
		return
	}

	sendResponse(w, bid)

    err =  saveBidToDB(bid)
	if err != nil {
		fmt.Println("DB error:", err)
		return
	} 
}

func fetchBidFromAPI() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), serverTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("request to api timed out after %v milliseconds", serverTimeout)
		}
		return "", fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	var quote QuoteResponse
	err = json.NewDecoder(resp.Body).Decode(&quote)
	if err != nil {
		fmt.Println("JSON error:", err)
		return "", err
	}
	return quote.USDBRL.Bid, nil
}

func saveBidToDB(bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), databaseTimeout)
	defer cancel()

	db, err := sql.Open("sqlite3", "./quotes.db")
	if err != nil {
    return fmt.Errorf("failed to open the SQLite database: %w", err)
}

	err = db.PingContext(ctx)
     if err != nil {
	 if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("request to database timed out after %v milliseconds", databaseTimeout)
	}
	return fmt.Errorf("failed to connect to database: %w", err)
}

    defer db.Close()

	stmt, err := db.Prepare("INSERT INTO quotes(value) VALUES(?)")

	if err != nil {
    fmt.Println("SQL prepare error:", err)
    return err
}

    defer stmt.Close()

	_, err = stmt.ExecContext(ctx, bid)
	if err != nil {
    return fmt.Errorf("failed to execute the insert statement for bid value '%s': %w", bid, err)
}
    return nil
}

func sendResponse(w http.ResponseWriter, bid string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"Dólar": bid})
}