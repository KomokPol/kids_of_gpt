package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PostJSON marshals payload to JSON, sends POST, returns raw response body and status code.
func PostJSON(client *http.Client, url string, payload any) ([]byte, int, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return out, resp.StatusCode, nil
}

// GetJSON sends GET and decodes JSON response into dst.
func GetJSON(client *http.Client, url string, dst any) (int, error) {
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return resp.StatusCode, fmt.Errorf("decode: %w", err)
	}
	return resp.StatusCode, nil
}
