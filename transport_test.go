package jsonrpc_client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPTransport(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected HTTP method: POST, got: %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json, got: %s", r.Header.Get("Content-Type"))
		}

		// Verify custom headers
		if r.Header.Get("X-API-Key") != "test-api-key" {
			t.Errorf("expected X-API-Key: test-api-key, got: %s", r.Header.Get("X-API-Key"))
		}

		// Read request body
		var req JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("request decode error: %v", err)
		}

		// Verify request
		if req.Version != "2.0" {
			t.Errorf("expected version: 2.0, got: %s", req.Version)
		}

		if req.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", req.Method)
		}

		// Create response
		resp := JSONRPCResponse{
			Version: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{"result":"success"}`),
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("response encode error: %v", err)
		}
	}))
	defer server.Close()

	// Create HTTP transport
	headers := map[string]string{
		"X-API-Key": "test-api-key",
	}
	transport := NewHTTPTransport(server.URL, headers)

	// Create request
	request := &JSONRPCRequest{
		Version: "2.0",
		ID:      NewID(1),
		Method:  "test.method",
		Params:  map[string]string{"key": "value"},
	}

	// Create response
	response := &JSONRPCResponse{
		ID: request.ID.New(),
	}

	// Send request
	err := transport.SendRequest(context.Background(), request, response)
	if err != nil {
		t.Fatalf("SendRequest error: %v", err)
	}

	// Verify response
	if response.Version != "2.0" {
		t.Errorf("expected version: 2.0, got: %s", response.Version)
	}

	if response.Error != nil {
		t.Errorf("error is not nil: %v", response.Error)
	}

	// Verify result
	var result map[string]string
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("result decode error: %v", err)
	}

	if result["result"] != "success" {
		t.Errorf("expected result: success, got: %s", result["result"])
	}
}

func TestHTTPTransportErrors(t *testing.T) {
	t.Run("JSON encode error", func(t *testing.T) {
		// Create a transport
		transport := NewHTTPTransport("http://example.com", nil)

		// Create a request that can't be marshaled to JSON
		// Use a function value which can't be marshaled to JSON
		fn := func() {}
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
			Params:  fn,
		}

		response := &JSONRPCResponse{
			ID: request.ID.New(),
		}

		err := transport.SendRequest(context.Background(), request, response)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var marshalErr *MarshalError
		if !errors.As(err, &marshalErr) {
			t.Fatalf("expected error type: *MarshalError, got: %T", err)
		}

		if marshalErr.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", marshalErr.Method)
		}
	})

	t.Run("http.NewRequestWithContext error", func(t *testing.T) {
		// Create a transport with an invalid URL that will cause NewRequestWithContext to fail
		transport := NewHTTPTransport("http://[::1]:namedport", nil)

		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		response := &JSONRPCResponse{
			ID: request.ID.New(),
		}

		err := transport.SendRequest(context.Background(), request, response)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var marshalErr *MarshalError
		if !errors.As(err, &marshalErr) {
			t.Fatalf("expected error type: *MarshalError, got: %T", err)
		}

		if marshalErr.Method != "test.method" {
			t.Errorf("expected method: test.method, got: %s", marshalErr.Method)
		}
	})

	t.Run("non-200 status code", func(t *testing.T) {
		// Test server that returns an error response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		transport := NewHTTPTransport(server.URL, nil)
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}
		response := &JSONRPCResponse{
			ID: request.ID.New(),
		}

		err := transport.SendRequest(context.Background(), request, response)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var statusErr *StatusCodeError
		if !errors.As(err, &statusErr) {
			t.Fatalf("expected error type: *StatusCodeError, got: %T", err)
		}

		if statusErr.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status code: %d, got: %d", http.StatusBadRequest, statusErr.StatusCode)
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		transport := NewHTTPTransport("invalid-url", nil)
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}
		response := &JSONRPCResponse{
			ID: request.ID.New(),
		}

		err := transport.SendRequest(context.Background(), request, response)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var invokeErr *InvokeError
		if !errors.As(err, &invokeErr) {
			t.Fatalf("expected error type: *InvokeError, got: %T", err)
		}
	})

	t.Run("JSON decode error", func(t *testing.T) {
		// Test server that returns invalid JSON
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		transport := NewHTTPTransport(server.URL, nil)
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}
		response := &JSONRPCResponse{
			ID: request.ID.New(),
		}

		err := transport.SendRequest(context.Background(), request, response)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var unmarshalErr *UnmarshalError
		if !errors.As(err, &unmarshalErr) {
			t.Fatalf("expected error type: *UnmarshalError, got: %T", err)
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		// Test server that delays response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sleep longer than the context timeout
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"result":"success"}}`))
		}))
		defer server.Close()

		transport := NewHTTPTransport(server.URL, nil)
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}
		response := &JSONRPCResponse{
			ID: request.ID.New(),
		}

		// Create a context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := transport.SendRequest(ctx, request, response)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var invokeErr *InvokeError
		if !errors.As(err, &invokeErr) {
			t.Fatalf("expected error type: *InvokeError, got: %T", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		// Test server that delays response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Sleep to allow time for cancellation
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"result":"success"}}`))
		}))
		defer server.Close()

		transport := NewHTTPTransport(server.URL, nil)
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}
		response := &JSONRPCResponse{
			ID: request.ID.New(),
		}

		// Create a context that will be canceled
		ctx, cancel := context.WithCancel(context.Background())

		// Start the request in a goroutine
		errCh := make(chan error, 1)
		go func() {
			errCh <- transport.SendRequest(ctx, request, response)
		}()

		// Cancel the context after a short delay
		time.Sleep(10 * time.Millisecond)
		cancel()

		// Wait for the result
		err := <-errCh
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var invokeErr *InvokeError
		if !errors.As(err, &invokeErr) {
			t.Fatalf("expected error type: *InvokeError, got: %T", err)
		}
	})
}

func TestHTTPTransportWithCustomClient(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"result":"success"}}`))
	}))
	defer server.Close()

	// Create a custom HTTP client with a timeout
	customClient := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	// Create HTTP transport with the custom client
	transport := &HTTPTransport{
		client:  customClient,
		baseURL: server.URL,
		headers: nil,
	}

	request := &JSONRPCRequest{
		Version: "2.0",
		ID:      NewID(1),
		Method:  "test.method",
	}
	response := &JSONRPCResponse{
		ID: request.ID.New(),
	}

	// Send request
	err := transport.SendRequest(context.Background(), request, response)
	if err != nil {
		t.Fatalf("SendRequest error: %v", err)
	}

	// Verify response
	var result map[string]string
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("result decode error: %v", err)
	}

	if result["result"] != "success" {
		t.Errorf("expected result: success, got: %s", result["result"])
	}
}
