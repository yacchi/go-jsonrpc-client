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
	transport := NewHTTPTransport(server.URL, WithHTTPHeaders(headers))

	// Create request
	request := &JSONRPCRequest{
		Version: "2.0",
		ID:      NewID(1),
		Method:  "test.method",
		Params:  map[string]string{"key": "value"},
	}

	// Create input and output
	input := &SendRequestInput{
		Requests: []*JSONRPCRequest{request},
		Batch:    false,
	}
	output := &SendRequestOutput{}

	// Send request
	var err error
	output, err = transport.SendRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("SendRequest error: %v", err)
	}

	// Verify response
	if len(output.Responses) == 0 {
		t.Fatalf("no response received")
	}
	response := output.Responses[0]

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

func TestHTTPTransportBatch(t *testing.T) {
	// Create a test HTTP server for batch requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "POST" {
			t.Errorf("expected HTTP method: POST, got: %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json, got: %s", r.Header.Get("Content-Type"))
		}

		// Read request body as array
		var reqs []*JSONRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			t.Fatalf("request decode error: %v", err)
		}

		// Verify we got a batch request
		if len(reqs) != 2 {
			t.Fatalf("expected 2 requests in batch, got: %d", len(reqs))
		}

		// Create batch response
		responses := []*JSONRPCResponse{
			{
				Version: "2.0",
				ID:      reqs[0].ID,
				Result:  json.RawMessage(`{"result":"success1"}`),
			},
			{
				Version: "2.0",
				ID:      reqs[1].ID,
				Result:  json.RawMessage(`{"result":"success2"}`),
			},
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			t.Fatalf("response encode error: %v", err)
		}
	}))
	defer server.Close()

	// Create HTTP transport
	transport := NewHTTPTransport(server.URL)

	// Create batch requests
	request1 := &JSONRPCRequest{
		Version: "2.0",
		ID:      NewID(1),
		Method:  "test.method1",
		Params:  map[string]string{"key": "value1"},
	}
	request2 := &JSONRPCRequest{
		Version: "2.0",
		ID:      NewID(2),
		Method:  "test.method2",
		Params:  map[string]string{"key": "value2"},
	}

	// Create input for batch request
	input := &SendRequestInput{
		Requests: []*JSONRPCRequest{request1, request2},
		Batch:    true,
	}

	// Send batch request
	output, err := transport.SendRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("SendRequest error: %v", err)
	}

	// Verify responses
	if len(output.Responses) != 2 {
		t.Fatalf("expected 2 responses, got: %d", len(output.Responses))
	}

	// Verify first response
	response1 := output.Responses[0]
	if response1.Version != "2.0" {
		t.Errorf("expected version: 2.0, got: %s", response1.Version)
	}
	if response1.Error != nil {
		t.Errorf("error is not nil: %v", response1.Error)
	}

	var result1 map[string]string
	if err := json.Unmarshal(response1.Result, &result1); err != nil {
		t.Fatalf("result decode error: %v", err)
	}
	if result1["result"] != "success1" {
		t.Errorf("expected result: success1, got: %s", result1["result"])
	}

	// Verify second response
	response2 := output.Responses[1]
	if response2.Version != "2.0" {
		t.Errorf("expected version: 2.0, got: %s", response2.Version)
	}
	if response2.Error != nil {
		t.Errorf("error is not nil: %v", response2.Error)
	}

	var result2 map[string]string
	if err := json.Unmarshal(response2.Result, &result2); err != nil {
		t.Fatalf("result decode error: %v", err)
	}
	if result2["result"] != "success2" {
		t.Errorf("expected result: success2, got: %s", result2["result"])
	}
}

func TestHTTPTransportErrors(t *testing.T) {
	t.Run("empty request list", func(t *testing.T) {
		// Create a transport
		transport := NewHTTPTransport("http://example.com")

		// Create empty input
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{},
			Batch:    false,
		}

		_, err := transport.SendRequest(context.Background(), input)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var invalidRequestErr *InvalidRequestError
		if !errors.As(err, &invalidRequestErr) {
			t.Fatalf("expected error type: *InvalidRequestError, got: %T", err)
		}

		if invalidRequestErr.Message != "no request provided" {
			t.Errorf("expected message: no request provided, got: %s", invalidRequestErr.Message)
		}
	})

	t.Run("JSON encode error", func(t *testing.T) {
		// Create a transport
		transport := NewHTTPTransport("http://example.com")

		// Create a request that can't be marshaled to JSON
		// Use a function value which can't be marshaled to JSON
		fn := func() {}
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
			Params:  fn,
		}

		// Create input
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}

		_, err := transport.SendRequest(context.Background(), input)
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

	t.Run("batch JSON encode error", func(t *testing.T) {
		// Create a transport
		transport := NewHTTPTransport("http://example.com")

		// Create a request that can't be marshaled to JSON
		// Use a function value which can't be marshaled to JSON
		fn := func() {}
		request1 := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method1",
			Params:  map[string]string{"key": "value"},
		}
		request2 := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(2),
			Method:  "test.method2",
			Params:  fn, // This will cause JSON marshaling to fail
		}

		// Create input for batch request with unmarshalable content
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request1, request2},
			Batch:    true,
		}

		_, err := transport.SendRequest(context.Background(), input)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var marshalErr *MarshalError
		if !errors.As(err, &marshalErr) {
			t.Fatalf("expected error type: *MarshalError, got: %T", err)
		}

		// The method should be from the first request in the batch
		if marshalErr.Method != "test.method1" {
			t.Errorf("expected method: test.method1, got: %s", marshalErr.Method)
		}
	})

	t.Run("http.NewRequestWithContext error", func(t *testing.T) {
		// Create a transport with an invalid URL that will cause NewRequestWithContext to fail
		transport := NewHTTPTransport("http://[::1]:namedport")

		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}

		_, err := transport.SendRequest(context.Background(), input)
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

		transport := NewHTTPTransport(server.URL)
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}

		_, err := transport.SendRequest(context.Background(), input)
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
		transport := NewHTTPTransport("invalid-url")
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}

		_, err := transport.SendRequest(context.Background(), input)
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

		transport := NewHTTPTransport(server.URL)
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}

		_, err := transport.SendRequest(context.Background(), input)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var unmarshalErr *UnmarshalError
		if !errors.As(err, &unmarshalErr) {
			t.Fatalf("expected error type: *UnmarshalError, got: %T", err)
		}
	})

	t.Run("batch JSON decode error", func(t *testing.T) {
		// Test server that returns invalid JSON for batch request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("invalid batch json"))
		}))
		defer server.Close()

		transport := NewHTTPTransport(server.URL)
		request1 := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method1",
		}
		request2 := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(2),
			Method:  "test.method2",
		}

		// Create input for batch request
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request1, request2},
			Batch:    true,
		}

		_, err := transport.SendRequest(context.Background(), input)
		if err == nil {
			t.Fatal("no error was returned")
		}

		// Verify error type
		var unmarshalErr *UnmarshalError
		if !errors.As(err, &unmarshalErr) {
			t.Fatalf("expected error type: *UnmarshalError, got: %T", err)
		}

		// Verify method name in error
		if unmarshalErr.Method != "test.method1" {
			t.Errorf("expected method: test.method1, got: %s", unmarshalErr.Method)
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

		transport := NewHTTPTransport(server.URL)
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}

		// Create a context with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := transport.SendRequest(ctx, input)
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

		transport := NewHTTPTransport(server.URL)
		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}

		// Create a context that will be canceled
		ctx, cancel := context.WithCancel(context.Background())

		// Start the request in a goroutine
		errCh := make(chan error, 1)
		go func() {
			_, err := transport.SendRequest(ctx, input)
			errCh <- err
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

	// Create input and output
	input := &SendRequestInput{
		Requests: []*JSONRPCRequest{request},
		Batch:    false,
	}
	output := &SendRequestOutput{}

	// Send request
	output, err := transport.SendRequest(context.Background(), input)
	if err != nil {
		t.Fatalf("SendRequest error: %v", err)
	}

	// Verify response
	if len(output.Responses) == 0 {
		t.Fatalf("no response received")
	}
	response := output.Responses[0]

	// Verify response
	var result map[string]string
	if err := json.Unmarshal(response.Result, &result); err != nil {
		t.Fatalf("result decode error: %v", err)
	}

	if result["result"] != "success" {
		t.Errorf("expected result: success, got: %s", result["result"])
	}
}

func TestHTTPTransportOptions(t *testing.T) {
	t.Run("WithHTTPClient", func(t *testing.T) {
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

		// Create HTTP transport with the custom client using WithHTTPClient option
		transport := NewHTTPTransport(server.URL, WithHTTPClient(customClient))

		// Verify that the custom client was set
		if transport.client != customClient {
			t.Errorf("expected client to be set to customClient")
		}

		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input and output
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}
		output := &SendRequestOutput{}

		// Send request
		output, err := transport.SendRequest(context.Background(), input)
		if err != nil {
			t.Fatalf("SendRequest error: %v", err)
		}

		// Verify response
		if len(output.Responses) == 0 {
			t.Fatalf("no response received")
		}
		response := output.Responses[0]

		// Verify response
		var result map[string]string
		if err := json.Unmarshal(response.Result, &result); err != nil {
			t.Fatalf("result decode error: %v", err)
		}

		if result["result"] != "success" {
			t.Errorf("expected result: success, got: %s", result["result"])
		}
	})

	t.Run("WithHTTPHeaders", func(t *testing.T) {
		// Create a test HTTP server that verifies headers
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify custom headers
			if r.Header.Get("X-API-Key") != "test-api-key" {
				t.Errorf("expected X-API-Key: test-api-key, got: %s", r.Header.Get("X-API-Key"))
			}
			if r.Header.Get("X-Custom-Header") != "custom-value" {
				t.Errorf("expected X-Custom-Header: custom-value, got: %s", r.Header.Get("X-Custom-Header"))
			}

			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"result":"success"}}`))
		}))
		defer server.Close()

		// Create custom headers
		headers := map[string]string{
			"X-API-Key":       "test-api-key",
			"X-Custom-Header": "custom-value",
		}

		// Create HTTP transport with custom headers using WithHTTPHeaders option
		transport := NewHTTPTransport(server.URL, WithHTTPHeaders(headers))

		// Verify that the headers were set
		if len(transport.headers) != 2 {
			t.Errorf("expected 2 headers, got: %d", len(transport.headers))
		}
		if transport.headers["X-API-Key"] != "test-api-key" {
			t.Errorf("expected X-API-Key: test-api-key, got: %s", transport.headers["X-API-Key"])
		}
		if transport.headers["X-Custom-Header"] != "custom-value" {
			t.Errorf("expected X-Custom-Header: custom-value, got: %s", transport.headers["X-Custom-Header"])
		}

		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input and output
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}
		output := &SendRequestOutput{}

		// Send request
		output, err := transport.SendRequest(context.Background(), input)
		if err != nil {
			t.Fatalf("SendRequest error: %v", err)
		}

		// Verify response
		if len(output.Responses) == 0 {
			t.Fatalf("no response received")
		}
		response := output.Responses[0]

		// Verify response
		var result map[string]string
		if err := json.Unmarshal(response.Result, &result); err != nil {
			t.Fatalf("result decode error: %v", err)
		}

		if result["result"] != "success" {
			t.Errorf("expected result: success, got: %s", result["result"])
		}
	})

	t.Run("Multiple options", func(t *testing.T) {
		// Create a test HTTP server that verifies headers
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify custom headers
			if r.Header.Get("X-API-Key") != "test-api-key" {
				t.Errorf("expected X-API-Key: test-api-key, got: %s", r.Header.Get("X-API-Key"))
			}

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

		// Create custom headers
		headers := map[string]string{
			"X-API-Key": "test-api-key",
		}

		// Create HTTP transport with both custom client and headers
		transport := NewHTTPTransport(
			server.URL,
			WithHTTPClient(customClient),
			WithHTTPHeaders(headers),
		)

		// Verify that both options were applied
		if transport.client != customClient {
			t.Errorf("expected client to be set to customClient")
		}
		if len(transport.headers) != 1 {
			t.Errorf("expected 1 header, got: %d", len(transport.headers))
		}
		if transport.headers["X-API-Key"] != "test-api-key" {
			t.Errorf("expected X-API-Key: test-api-key, got: %s", transport.headers["X-API-Key"])
		}

		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input and output
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}
		output := &SendRequestOutput{}

		// Send request
		output, err := transport.SendRequest(context.Background(), input)
		if err != nil {
			t.Fatalf("SendRequest error: %v", err)
		}

		// Verify response
		if len(output.Responses) == 0 {
			t.Fatalf("no response received")
		}
		response := output.Responses[0]

		// Verify response
		var result map[string]string
		if err := json.Unmarshal(response.Result, &result); err != nil {
			t.Fatalf("result decode error: %v", err)
		}

		if result["result"] != "success" {
			t.Errorf("expected result: success, got: %s", result["result"])
		}
	})

	t.Run("No options", func(t *testing.T) {
		// Create a test HTTP server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Send response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"result":"success"}}`))
		}))
		defer server.Close()

		// Create HTTP transport with no options
		transport := NewHTTPTransport(server.URL)

		// Verify default client was created
		if transport.client == nil {
			t.Errorf("expected default client to be created")
		}
		if transport.headers != nil {
			t.Errorf("expected headers to be nil, got: %v", transport.headers)
		}

		request := &JSONRPCRequest{
			Version: "2.0",
			ID:      NewID(1),
			Method:  "test.method",
		}

		// Create input and output
		input := &SendRequestInput{
			Requests: []*JSONRPCRequest{request},
			Batch:    false,
		}
		output := &SendRequestOutput{}

		// Send request
		output, err := transport.SendRequest(context.Background(), input)
		if err != nil {
			t.Fatalf("SendRequest error: %v", err)
		}

		// Verify response
		if len(output.Responses) == 0 {
			t.Fatalf("no response received")
		}
		response := output.Responses[0]

		// Verify response
		var result map[string]string
		if err := json.Unmarshal(response.Result, &result); err != nil {
			t.Fatalf("result decode error: %v", err)
		}

		if result["result"] != "success" {
			t.Errorf("expected result: success, got: %s", result["result"])
		}
	})
}
