package exchange

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRate_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/convert?from=USD&to=EUR", r.URL.String())
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"base":"USD","target":"EUR","rate":0.85}`))
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL)
	rate, err := svc.GetRate(context.Background(), "USD", "EUR")

	require.NoError(t, err)
	assert.Equal(t, 0.85, rate)
}

func TestGetRate_APIBusinessError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid currency pair"}`))
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL)
	_, err := svc.GetRate(context.Background(), "USD", "UNKNOWN")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid currency pair")
}

func TestGetRate_MalformedJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`this is not json`))
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL)
	_, err := svc.GetRate(context.Background(), "USD", "EUR")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode error")
}

func TestGetRate_SlowResponseOrTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Millisecond) // simulates slow response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"base":"USD","target":"EUR","rate":0.85}`))
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL)
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond) // Will timeout
	defer cancel()

	_, err := svc.GetRate(ctx, "USD", "EUR")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestGetRate_ServerPanic_500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL)
	_, err := svc.GetRate(context.Background(), "USD", "EUR")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status: 500")
}

func TestGetRate_EmptyBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL)
	_, err := svc.GetRate(context.Background(), "USD", "EUR")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "decode error: EOF")
}

func TestGetRate_Retry(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"base":"USD","target":"EUR","rate":0.85}`))
	}))
	defer ts.Close()

	svc := NewExchangeService(ts.URL)
	rate, err := svc.GetRate(context.Background(), "USD", "EUR")

	require.NoError(t, err)
	assert.Equal(t, 0.85, rate)
	assert.Equal(t, 2, attempts)
}
