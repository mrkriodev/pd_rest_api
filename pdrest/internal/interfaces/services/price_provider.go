package services

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Default Pyth mainnet price feed IDs (BASE/USD); USDT-quoted pairs use the same USD feed.
// See https://docs.pyth.network/price-feeds/pro/price-feed-ids
var pythFeedIDs = map[string]string{
	"ETH/USDT": "0xff61491a931112ddf1bd8147cd1b641375f79f5825126d665480874634fd0ace",
	"ETH/USD":  "0xff61491a931112ddf1bd8147cd1b641375f79f5825126d665480874634fd0ace",
	"BTC/USDT": "0xe62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43",
	"BTC/USD":  "0xe62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43",
	"SOL/USDT": "0xef0d8b6fda2ceba41da15d4095d1da392a0d2f8ed0c6c7bc0f4cfac8c280b56d",
	"SOL/USD":  "0xef0d8b6fda2ceba41da15d4095d1da392a0d2f8ed0c6c7bc0f4cfac8c280b56d",
}

type PriceProvider struct {
	baseURL string
	client  *http.Client
}

func NewPriceProvider(baseURL string) *PriceProvider {
	if baseURL == "" {
		baseURL = "https://hermes.pyth.network"
	}
	return &PriceProvider{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type pythPricePayload struct {
	Price string `json:"price"`
	Conf  string `json:"conf"`
	Expo  int    `json:"expo"`
}

type pythPriceFeedItem struct {
	ID    string           `json:"id"`
	Price pythPricePayload `json:"price"`
}

// GetPrice fetches the latest aggregate price for a trading pair via Pyth Hermes
// (GET /api/latest_price_feeds). Pair format: "ETH/USDT" (maps to ETH/USD feed).
func (p *PriceProvider) GetPrice(pair string) (float64, error) {
	key := strings.ToUpper(strings.TrimSpace(pair))
	feedID, ok := pythFeedIDs[key]
	if !ok {
		return 0, fmt.Errorf("no Pyth feed ID configured for pair %q", pair)
	}

	u, err := url.Parse(p.baseURL + "/api/latest_price_feeds")
	if err != nil {
		return 0, fmt.Errorf("invalid Hermes base URL: %w", err)
	}
	q := u.Query()
	q.Add("ids[]", feedID)
	u.RawQuery = q.Encode()

	resp, err := p.client.Get(u.String())
	if err != nil {
		return 0, fmt.Errorf("failed to fetch price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("price provider returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read price response: %w", err)
	}

	var items []pythPriceFeedItem
	if err := json.Unmarshal(body, &items); err != nil {
		return 0, fmt.Errorf("failed to decode price response: %w", err)
	}
	if len(items) == 0 || items[0].Price.Price == "" {
		return 0, fmt.Errorf("price provider returned no price data")
	}

	priceInt, err := strconv.ParseInt(items[0].Price.Price, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse price integer: %w", err)
	}

	expo := items[0].Price.Expo
	value := float64(priceInt) * math.Pow10(expo)
	return value, nil
}
