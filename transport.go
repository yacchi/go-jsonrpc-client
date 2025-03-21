package jsonrpc_client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// SendRequestInput represents input parameters for sending a request
type SendRequestInput struct {
	Requests []*JSONRPCRequest
	Batch    bool
}

// SendRequestOutput represents output results of sending a request
type SendRequestOutput struct {
	Responses []*JSONRPCResponse
}

// Transport is an interface for sending JSON-RPC requests
type Transport interface {
	// SendRequest sends a JSON-RPC request and returns the response
	SendRequest(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error)
}

// HTTPTransport is a transport for sending JSON-RPC requests via HTTP
type HTTPTransport struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

type HTTPTransportOption func(*HTTPTransport)

// WithHTTPClient sets the HTTP client for the transport
func WithHTTPClient(client *http.Client) HTTPTransportOption {
	return func(t *HTTPTransport) {
		t.client = client
	}
}

// WithHTTPHeaders sets the HTTP headers for the transport
func WithHTTPHeaders(headers map[string]string) HTTPTransportOption {
	return func(t *HTTPTransport) {
		t.headers = headers
	}
}

// NewHTTPTransport creates a transport for sending JSON-RPC requests via HTTP
func NewHTTPTransport(baseURL string, opts ...HTTPTransportOption) *HTTPTransport {
	t := &HTTPTransport{
		client:  &http.Client{},
		baseURL: baseURL,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// SendRequest sends a JSON-RPC request via HTTP
func (t *HTTPTransport) SendRequest(ctx context.Context, input *SendRequestInput) (*SendRequestOutput, error) {
	if len(input.Requests) == 0 {
		return nil, &InvalidRequestError{Message: "no request provided"}
	}

	method := input.Requests[0].Method
	body := bytes.NewBuffer(nil)

	if input.Batch {
		if err := json.NewEncoder(body).Encode(input.Requests); err != nil {
			return nil, &MarshalError{Method: method, Err: err}
		}
	} else {
		if err := json.NewEncoder(body).Encode(input.Requests[0]); err != nil {
			return nil, &MarshalError{Method: method, Err: err}
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, body)
	if err != nil {
		return nil, &MarshalError{Method: method, Err: err}
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, &InvokeError{Method: method, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &StatusCodeError{Method: method, StatusCode: resp.StatusCode}
	}

	output := &SendRequestOutput{}

	if input.Batch {
		// Decode batch response
		if err := json.NewDecoder(resp.Body).Decode(&output.Responses); err != nil {
			return nil, &UnmarshalError{Method: method, Err: err}
		}
	} else {
		// Process single request
		var response *JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, &UnmarshalError{Method: method, Err: err}
		}
		output.Responses = []*JSONRPCResponse{response}
	}

	return output, nil
}
