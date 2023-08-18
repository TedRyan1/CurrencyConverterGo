package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	serverURL        = "http://localhost:8080/cotacao"
	clientTimeout    = 300 * time.Millisecond
	cotacaoFileName  = "cotacao.txt"
)

func main() {
	bid, err := fetchBidFromServer()
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}

	err = saveBidToFile(bid)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("Cotação saved to", cotacaoFileName, "!")
}

func fetchBidFromServer() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request to server timed out after %v milliseconds", clientTimeout)
		}
		return nil, fmt.Errorf("error making request to server: %w", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return nil, err
	}
	return body, nil
}

func saveBidToFile(bid []byte) error {
	return os.WriteFile(cotacaoFileName, bid, 0644)
}