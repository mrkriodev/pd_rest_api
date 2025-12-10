package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type PriceProvider struct {
	baseURL string
	client  *http.Client
}

func NewPriceProvider(baseURL string) *PriceProvider {
	if baseURL == "" {
		baseURL = "https://api.binance.com/api/v3/ticker/price"
	}
	return &PriceProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type BinancePriceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// GetPrice fetches the current price for a trading pair
// pair format: "ETH/USDT" -> converts to "ETHUSDT" for Binance
func (p *PriceProvider) GetPrice(pair string) (float64, error) {
	// Convert pair format from "ETH/USDT" to "ETHUSDT"
	symbol := strings.ReplaceAll(strings.ToUpper(pair), "/", "")

	// Build URL - for Binance API: /api/v3/ticker/price?symbol=ETHUSDT
	url := fmt.Sprintf("%s?symbol=%s", p.baseURL, symbol)

	resp, err := p.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("price provider returned status %d: %s", resp.StatusCode, string(body))
	}

	var priceResp BinancePriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return 0, fmt.Errorf("failed to decode price response: %w", err)
	}

	var price float64
	if _, err := fmt.Sscanf(priceResp.Price, "%f", &price); err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	return price, nil
}
