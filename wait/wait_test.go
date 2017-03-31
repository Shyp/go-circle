package wait

import (
	"io"
	"net/http"
	"testing"
	"time"
)

func makeRequest(client http.Client, method, uri string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", "wait-test")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func TestHttpError(t *testing.T) {
	client := http.Client{
		Timeout: 200 * time.Millisecond,
	}
	_, err := makeRequest(client, "GET", "http://localhost:11233")
	if !isHttpError(err) {
		t.Fatalf("expected err to be http error, was %s", err)
	}

	_, err = makeRequest(client, "GET", "https://httpbin.org/delay/2")
	if !isHttpError(err) {
		t.Fatalf("expected err to be http error, was %s", err)
	}
}

func TestEffectiveCost(t *testing.T) {
	cost := getEffectiveCost(1 * time.Hour)
	if cost != 7897 {
		t.Errorf("expected 1 hour cost to be %d, was %d", 7897, cost)
	}
	cost = getEffectiveCost(30 * time.Minute)
	if cost != 3948 {
		t.Errorf("expected half hour cost to be %d, was %d", 3948, cost)
	}
	cost = getEffectiveCost(2 * time.Hour)
	if cost != 15793 {
		t.Errorf("expected half hour cost to be %d, was %d", 15793, cost)
	}
}

func TestGetMaxTipLength(t *testing.T) {
	minTipLength := getMinTipLength("1d79f2b877c86ac0964f3fe69a0171926aa6f1d8", "1d79f2b87")
	expectedMinTipLength := 9
	if minTipLength != expectedMinTipLength {
		t.Errorf("expected half hour cost to be %d, was %d", expectedMinTipLength, minTipLength)
	}

	minTipLength = getMinTipLength("1d79f2b877c86ac0964f3fe69a0171926aa6f1d8", "1d79f2b")
	expectedMinTipLength = 7
	if minTipLength != expectedMinTipLength {
		t.Errorf("expected half hour cost to be %d, was %d", expectedMinTipLength, minTipLength)
	}

	minTipLength = getMinTipLength("1d79f", "1d79f2b87")
	expectedMinTipLength = 5
	if minTipLength != expectedMinTipLength {
		t.Errorf("expected half hour cost to be %d, was %d", expectedMinTipLength, minTipLength)
	}
}
