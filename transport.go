package jsonrpc_client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// Transport is an interface for sending JSON-RPC requests
type Transport interface {
	// SendRequest sends a JSON-RPC request and returns the response
	// Returns request payload, response payload, and error
	SendRequest(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error
}

// HTTPTransport is a transport for sending JSON-RPC requests via HTTP
type HTTPTransport struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// NewHTTPTransport creates a transport for sending JSON-RPC requests via HTTP
func NewHTTPTransport(baseURL string, headers map[string]string) *HTTPTransport {
	return &HTTPTransport{
		client:  &http.Client{},
		baseURL: baseURL,
		headers: headers,
	}
}

// SendRequest sends a JSON-RPC request via HTTP
func (t *HTTPTransport) SendRequest(ctx context.Context, request *JSONRPCRequest, response *JSONRPCResponse) error {
	method := request.Method

	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(request); err != nil {
		return &MarshalError{Method: method, Err: err}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, body)
	if err != nil {
		return &MarshalError{Method: method, Err: err}
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return &InvokeError{Method: method, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &StatusCodeError{Method: method, StatusCode: int32(resp.StatusCode)}
	}

	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return &UnmarshalError{Method: method, Err: err}
	}
	return nil
}
