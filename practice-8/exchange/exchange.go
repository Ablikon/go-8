package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RateResponse struct {
	Base     string  `json:"base"`
	Target   string  `json:"target"`
	Rate     float64 `json:"rate"`
	ErrorMsg string  `json:"error,omitempty"`
}

type ExchangeService struct {
	BaseURL string
	Client  *http.Client
}

func NewExchangeService(baseURL string) *ExchangeService {
	return &ExchangeService{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetRate requests the rate. Example URL: /convert?from=USD&to=EUR
func (s *ExchangeService) GetRate(ctx context.Context, from, to string) (float64, error) {
	url := fmt.Sprintf("%s/convert?from=%s&to=%s", s.BaseURL, from, to)

	var lastErr error
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return 0, fmt.Errorf("create request error: %w", err)
		}

		var attemptRate float64
		var attemptErr error

		func() {
			resp, doErr := s.Client.Do(req)
			if doErr != nil {
				attemptErr = fmt.Errorf("network error: %w", doErr)
				return
			}
			defer resp.Body.Close()

			var result RateResponse
			if resp.StatusCode != http.StatusOK {
				if err := json.NewDecoder(resp.Body).Decode(&result); err == nil && result.ErrorMsg != "" {
					attemptErr = fmt.Errorf("api error: %s", result.ErrorMsg)
				} else {
					attemptErr = fmt.Errorf("unexpected status: %d", resp.StatusCode)
				}
				return
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				attemptErr = fmt.Errorf("decode error: %w", err)
				return
			}

			attemptRate = result.Rate
		}()

		if attemptErr != nil {
			lastErr = attemptErr
			time.Sleep(10 * time.Millisecond) // short delay before retry
			continue
		}

		return attemptRate, nil
	}

	return 0, fmt.Errorf("failed after 3 attempts, last error: %w", lastErr)
}
