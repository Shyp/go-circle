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
