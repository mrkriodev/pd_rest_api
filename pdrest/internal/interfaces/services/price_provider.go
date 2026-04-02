package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PriceProvider struct {
	baseURL string
	client  *http.Client
}

func NewPriceProvider(baseURL string) *PriceProvider {
	if baseURL == "" {
		baseURL = "https://www.okx.com/api/v5/market/candles"
	}
	return &PriceProvider{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type OKXCandlesResponse struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data [][]string `json:"data"`
}

// GetPrice fetches the current price for a trading pair
// pair format: "ETH/USDT" -> converts to "ETH-USDT" for OKX instId
func (p *PriceProvider) GetPrice(pair string) (float64, error) {
	// Convert pair format from "ETH/USDT" to "ETH-USDT"
	instID := strings.ReplaceAll(strings.ToUpper(pair), "/", "-")

	// Build URL - OKX candles API
	url := fmt.Sprintf("%s?instId=%s&bar=1s&limit=300", p.baseURL, instID)

	resp, err := p.client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("price provider returned status %d: %s", resp.StatusCode, string(body))
	}

	var priceResp OKXCandlesResponse
	if err := json.NewDecoder(resp.Body).Decode(&priceResp); err != nil {
		return 0, fmt.Errorf("failed to decode price response: %w", err)
	}
	if priceResp.Code != "0" {
		return 0, fmt.Errorf("price provider returned code %s: %s", priceResp.Code, priceResp.Msg)
	}
	if len(priceResp.Data) == 0 || len(priceResp.Data[0]) < 5 {
		return 0, fmt.Errorf("price provider returned empty candles data")
	}

	// OKX candle format: [ts, open, high, low, close, ...]
	price, err := strconv.ParseFloat(priceResp.Data[0][4], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	return price, nil
}
